package bill

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/pay"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util/log"
)

// Config is a billing monitor configuration.
type Config struct {
	CollectPeriod  uint // In milliseconds.
	RequestTLS     bool
	RequestTimeout uint // In milliseconds, must be less than CollectPeriod.
}

// NewConfig creates a new billing monitor configuration.
func NewConfig() *Config {
	return &Config{
		CollectPeriod:  5000,
		RequestTLS:     false,
		RequestTimeout: 2500,
	}
}

type postChequeFunc func(db *reform.DB, channel, pscAddr, pass string,
	amount uint64, tls bool, timeout uint, pr *proc.Processor) error

// Monitor is a client billing monitor.
type Monitor struct {
	conf   *Config
	logger log.Logger
	db     *reform.DB
	pr     *proc.Processor
	psc    string
	pw     data.PWDGetter
	post   postChequeFunc // Is overrided in unit-tests.
	mtx    sync.Mutex     // To guard the exit channels.
	exit   chan struct{}
	exited chan struct{}
	// The channel is only needed for tests.
	// It allows to get a result of a processing.
	processErrors chan error
	// The channel is only needed for tests.
	// It allows to get a result of a posting of cheques.
	postChequeErrors chan error
}

// NewMonitor creates a new client billing monitor.
func NewMonitor(conf *Config, logger log.Logger, db *reform.DB,
	pr *proc.Processor, pscAddr string, pw data.PWDGetter) *Monitor {
	return &Monitor{
		conf:   conf,
		logger: logger.Add("type", "client/bill.Monitor"),
		db:     db,
		pr:     pr,
		psc:    pscAddr,
		pw:     pw,
		post:   pay.PostCheque,
	}
}

// Run processes billing for active client channels. This function does not
// return until an error occurs or Close() is called.
func (m *Monitor) Run() error {
	m.mtx.Lock()
	if m.exit != nil {
		m.mtx.Unlock()
		return ErrAlreadyRunning
	}
	m.exit = make(chan struct{}, 1)
	m.exited = make(chan struct{}, 1)
	m.mtx.Unlock()

	period := time.Duration(m.conf.CollectPeriod) * time.Millisecond
L:
	for {
		select {
		case <-m.exit:
			break L
		default:
		}

		started := time.Now()

		chans, err := m.db.SelectAllFrom(data.ChannelTable, `
			 JOIN accounts ON eth_addr = client
			WHERE service_status IN ('active', 'suspended')
			  AND channel_status = 'active' AND in_use`)
		if err != nil {
			m.logger.Error(err.Error())
			break L
		}

		for _, v := range chans {
			err := m.processChannel(v.(*data.Channel))

			select {
			case m.processErrors <- err:
			default:
			}

			if err != nil {
				break L
			}
		}

		time.Sleep(period - time.Now().Sub(started))
	}

	m.exited <- struct{}{}

	m.mtx.Lock()
	m.exit = nil
	m.mtx.Unlock()

	m.logger.Info(ErrMonitorClosed.Error())
	return ErrMonitorClosed
}

// Close causes currently running Run() function to exit.
func (m *Monitor) Close() {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if m.exit != nil {
		m.exit <- struct{}{}
		<-m.exited
	}
}

func (m *Monitor) processChannel(ch *data.Channel) error {
	var offer data.Offering
	if err := m.db.FindByPrimaryKeyTo(&offer, ch.Offering); err != nil {
		m.logger.Add("offering", ch.Offering).Error(err.Error())
		return ErrGetOffering
	}

	terminate, err := m.isToBeTerminated(ch, &offer)
	if err != nil {
		return err
	}
	if terminate {
		_, err := m.pr.TerminateChannel(ch.ID, data.JobBillingChecker, false)
		if err != nil {
			if err != proc.ErrSameJobExists {
				m.logger.Add("channel",
					ch.ID).Error(err.Error())
				return err
			}
			m.logger.Add("channel", ch.ID, "error",
				err.Error()).Debug("failed to trigger" +
				" termination")
		} else {
			m.logger.Add("channel", ch.ID).Info("trigger" +
				" termination")
		}
		return nil
	}

	var consumed uint64
	if err := m.db.QueryRow(`
		SELECT COALESCE(sum(units_used),0)
		  FROM sessions
		 WHERE channel = $1`, ch.ID).Scan(&consumed); err != nil {
		m.logger.Add("channel", ch.ID).Error(err.Error())
		return ErrGetConsumedUnits
	}

	lag := int64(consumed)/int64(offer.BillingInterval) -
		(int64(ch.ReceiptBalance)-int64(offer.SetupPrice))/
			int64(offer.UnitPrice)
	if lag <= 0 {
		return nil
	}

	amount := consumed*offer.UnitPrice + offer.SetupPrice
	if amount > ch.TotalDeposit {
		amount = ch.TotalDeposit
	}

	go m.postCheque(ch.ID, amount)

	return nil
}

func (m *Monitor) isToBeTerminated(
	ch *data.Channel, offer *data.Offering) (bool, error) {

	if ch.ReceiptBalance == ch.TotalDeposit {
		m.logger.Debug("channel is complete")
		return true, nil
	}

	m.logger.Debug("channel not complete")

	reached, err := m.maxInactiveTimeReached(ch, offer)
	if err != nil {
		m.logger.Error(err.Error())
		// TODO: add error
		return false, ErrGetConsumedUnits
	}
	return reached, nil
}

func (m *Monitor) maxInactiveTimeReached(
	ch *data.Channel, offer *data.Offering) (bool, error) {
	if offer.MaxInactiveTimeSec == nil {
		return false, nil
	}
	query := "SELECT COUNT(*), MAX(last_usage_time) from sessions"
	var lastUsage time.Time
	var qty uint
	if err := m.db.QueryRow(query).Scan(&qty, &lastUsage); err != nil {
		return false, err
	}
	if qty == 0 {
		lastUsage = ch.PreparedAt
	}
	inactiveSeconds := uint64(time.Since(lastUsage).Seconds())
	return qty > 0 && inactiveSeconds > *offer.MaxInactiveTimeSec, nil
}

func (m *Monitor) postCheque(ch string, amount uint64) {
	m.logger.Add("amount", amount).Info("posting cheque")
	handleErr := func(err error) {
		select {
		case m.postChequeErrors <- err:
		default:
		}
	}

	pscHex := data.HexFromBytes(common.HexToAddress(m.psc).Bytes())
	err := m.post(m.db, ch, pscHex, m.pw.Get(), amount,
		m.conf.RequestTLS, m.conf.RequestTimeout, m.pr)
	if err != nil {
		m.logger.Add("channel", ch, "amount",
			amount).Error(err.Error())
		go handleErr(err)
		return
	}

	res, err := m.db.Exec(`
		UPDATE channels
		   SET receipt_balance = $1
		 WHERE id = $2 AND receipt_balance < $1`, amount, ch)
	if err != nil {
		m.logger.Add("channel", ch, "amount",
			amount).Error(ErrUpdateReceiptBalance.Error())
		go handleErr(err)
		return
	}

	n, err := res.RowsAffected()
	if err != nil {
		if n != 0 {
			m.logger.Add("channel", ch, "amount",
				amount).Info("updated receipt balance")
		} else {
			m.logger.Add("channel", ch).Warn(
				"receipt balance isn't updated")
		}
	}
	go handleErr(err)
}
