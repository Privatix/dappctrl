package bill

import (
	"errors"
	"sync"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/pay"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util"
)

// Errors.
var (
	ErrAlreadyRunning = errors.New("already running")
	ErrMonitorClosed  = errors.New("monitor closed")
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
	amount uint64, tls bool, timeout uint) error

// Monitor is a client billing monitor.
type Monitor struct {
	conf   *Config
	logger *util.Logger
	db     *reform.DB
	pr     *proc.Processor
	psc    string
	pw     data.PWDGetter
	post   postChequeFunc // Is overrided in unit-tests.
	mtx    sync.Mutex     // To guard the exit channels.
	exit   chan struct{}
	exited chan struct{}
}

// NewMonitor creates a new client billing monitor.
func NewMonitor(conf *Config, logger *util.Logger, db *reform.DB,
	pr *proc.Processor, pscAddr string, pw data.PWDGetter) *Monitor {
	return &Monitor{
		conf:   conf,
		logger: logger,
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
	ret := ErrMonitorClosed
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
			ret = err
			break L
		}

		for _, v := range chans {
			err := m.processChannel(v.(*data.Channel))
			if err != nil {
				ret = err
				break L
			}
		}

		time.Sleep(period - time.Now().Sub(started))
	}

	m.exited <- struct{}{}

	m.mtx.Lock()
	m.exit = nil
	m.mtx.Unlock()

	logger.Info("%s", ret)
	return ret
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
	if ch.ReceiptBalance == ch.TotalDeposit {
		_, err := m.pr.TerminateChannel(ch.ID, data.JobBillingChecker)
		if err != nil {
			if err != proc.ErrSameJobExists {
				return err
			}
			msg := "failed to trigger termination for chan %s: %s"
			m.logger.Error(msg, ch.ID, err)
		} else {
			msg := "triggered termination for chan %s"
			m.logger.Info(msg, ch.ID)
		}
		return nil
	}

	var consumed uint64
	if err := m.db.QueryRow(`
		SELECT sum(units_used)
		  FROM sessions
		 WHERE channel = $1`, ch.ID).Scan(&consumed); err != nil {
		return err
	}

	var offer data.Offering
	if err := m.db.FindByPrimaryKeyTo(&offer, ch.Offering); err != nil {
		return err
	}

	lag := int64(consumed)/int64(offer.BillingInterval) -
		(int64(ch.ReceiptBalance)-int64(offer.SetupPrice))/
			int64(offer.UnitPrice)
	if lag <= 0 {
		return nil
	}

	amount := consumed/offer.UnitPrice + offer.SetupPrice
	if amount > ch.TotalDeposit {
		amount = ch.TotalDeposit
	}

	go m.postCheque(ch.ID, amount)

	return nil
}

func (m *Monitor) postCheque(ch string, amount uint64) {
	err := m.post(m.db, ch, m.psc, m.pw.Get(), amount,
		m.conf.RequestTLS, m.conf.RequestTimeout)
	if err != nil {
		m.logger.Error("failed to post cheque for chan %s: %s", ch, err)
		return
	}

	res, err := m.db.Exec(`
		UPDATE channels
		   SET receipt_balance = $1
		 WHERE id = $2 AND receipt_balance < $1`, amount, ch)
	if err != nil {
		msg := "failed to update receipt balance for chan %s: %s"
		m.logger.Error(msg, ch, err)
		return
	}

	if n, err := res.RowsAffected(); err != nil {
		if n != 0 {
			msg := "updated receipt balance for chan %s: %d"
			m.logger.Info(msg, ch, amount)
		} else {
			msg := "receipt balance isn't updated for chan %s"
			m.logger.Warn(msg, ch)
		}
	}
}
