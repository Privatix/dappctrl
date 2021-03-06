package bill

import (
	"context"
	"crypto/ecdsa"
	"database/sql"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lib/pq"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/pay"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/util/srv"
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

type postChequeFunc func(db *reform.DB, channel *data.Channel,
	pscAddr data.HexString, key *ecdsa.PrivateKey, amount uint64,
	tls bool, timeout uint, pr *proc.Processor) error

// PriceSuggestor suggests best gas price for current moment.
type PriceSuggestor interface {
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
}

// Monitor is a client billing monitor.
type Monitor struct {
	conf      *Config
	logger    log.Logger
	db        *reform.DB
	pr        *proc.Processor
	queue     job.Queue
	psc       string
	pw        data.PWDGetter
	post      postChequeFunc // Is overrided in unit-tests.
	mtx       sync.Mutex     // To guard the exit channels.
	suggestor PriceSuggestor
	exit      chan struct{}
	exited    chan struct{}
	// The channel is only needed for tests.
	// It allows to get a result of a processing.
	processErrors chan error
	// The channel is only needed for tests.
	// It allows to get a result of a posting of cheques.
	postChequeErrors chan error
	// Auto increase deposit enabled/disabled. Read from settings table.
	autoIncrease bool
	// Auto increase deposit when used percent of all available traffic.
	// Read from settings table.
	autoIncreaseAtRate float64
}

// NewMonitor creates a new client billing monitor.
func NewMonitor(conf *Config, logger log.Logger, db *reform.DB, suggestor PriceSuggestor,
	pr *proc.Processor, queue job.Queue, pscAddr string, pw data.PWDGetter) *Monitor {
	return &Monitor{
		conf:      conf,
		logger:    logger.Add("type", "client/bill.Monitor"),
		db:        db,
		pr:        pr,
		queue:     queue,
		psc:       pscAddr,
		pw:        pw,
		post:      pay.PostCheque,
		suggestor: suggestor,
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

		if err := m.readSettings(); err != nil {
			m.logger.Error(err.Error())
			break L
		}

		for _, v := range chans {
			err = m.processChannel(v.(*data.Channel))

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

func (m *Monitor) readSettings() (err error) {
	m.autoIncrease, err = data.ReadBoolSetting(m.db.Querier, data.SettingClientAutoincreaseDeposit)
	if err != nil {
		return err
	}
	percent, err := data.ReadUintSetting(m.db.Querier, data.SettingClientAutoincreaseDepositPercent)
	if err != nil {
		return err
	}
	m.autoIncreaseAtRate = float64(percent) / 100
	return nil
}

func (m *Monitor) processChannel(ch *data.Channel) error {
	logger := m.logger.Add("method", "processChannel", "channel", ch)
	var offer data.Offering
	if err := m.db.FindByPrimaryKeyTo(&offer, ch.Offering); err != nil {
		logger.Error(err.Error())
		return ErrGetOffering
	}

	terminate, err := m.isToBeTerminated(logger, ch, &offer)
	if err != nil {
		return err
	}
	if terminate {
		_, err := m.pr.TerminateChannel(ch.ID,
			data.JobBillingChecker, false)
		if err != nil {
			if err != proc.ErrSameJobExists {
				logger.Error(err.Error())
				return err
			}
			logger.Add("error", err.Error()).Debug(
				"failed to trigger termination")
		} else {
			logger.Info("trigger termination")
		}
		return nil
	}

	var consumed uint64
	if err := m.db.QueryRow(`
		SELECT COALESCE(sum(units_used),0)
		  FROM sessions
		 WHERE channel = $1`, ch.ID).Scan(&consumed); err != nil {
		logger.Error(err.Error())
		return ErrGetConsumedUnits
	}

	amount := data.ComputePrice(&offer, consumed)
	if amount > ch.TotalDeposit {
		amount = ch.TotalDeposit
		go m.postCheque(ch.ID, amount)
		return nil
	}

	if err := m.checkAndCreateAutoIncreaseJob(logger, ch, amount); err != nil {
		logger.Error(err.Error())
		return err
	}

	lag := int64(consumed) - (int64(ch.ReceiptBalance)-
		int64(offer.SetupPrice))/int64(offer.UnitPrice)
	if lag/int64(offer.BillingInterval) >= 1 {
		go m.postCheque(ch.ID, amount)
	}

	return nil
}

func (m *Monitor) checkAndCreateAutoIncreaseJob(logger log.Logger, ch *data.Channel, amount uint64) error {
	autoincreaseAfter := uint64(float64(ch.TotalDeposit) * m.autoIncreaseAtRate)
	if !m.autoIncrease || amount < autoincreaseAfter {
		return nil
	}
	err, exists := notMinedChannelTopUpExists(logger, m.db, ch.ID)
	if exists {
		logger.Debug("active channel topup exists")
		return nil
	}
	if err != nil {
		logger.Error(err.Error())
		return nil
	}
	suggestedGasPrice, err := m.fetchSuggestedGasPrice()
	if err != nil {
		return fmt.Errorf("could not fetch suggested gas price: %v", err)
	}
	logger = logger.Add("suggestedGasPrice", suggestedGasPrice.Uint64())
	jdata := &data.JobTopUpChannelData{
		GasPrice: suggestedGasPrice.Uint64(),
		Deposit:  ch.TotalDeposit,
	}
	err = job.AddWithData(m.queue, nil, data.JobClientPreChannelTopUp, data.JobChannel, ch.ID, data.JobBillingChecker, jdata)
	if err == job.ErrAlreadyProcessing || err == job.ErrDuplicatedJob {
		logger.Warn("active channel topup exists")
		return nil
	}
	return err
}

func notMinedChannelTopUpExists(logger log.Logger, db *reform.DB, chID string) (error, bool) {
	// If it's first time
	if err := db.SelectOneTo(&data.Job{}, "WHERE related_id=$1 AND type=$2",
		chID, data.JobClientPreChannelTopUp); err == sql.ErrNoRows {
		return nil, false
	}
	// If there's pending jobs or completed but waiting for after topup job.
	row := db.QueryRow(`
		SELECT count(*)
		  FROM jobs
		 WHERE status in ($1, $2)
			   AND type = $3
			   AND related_id=$5
			   AND created_at > (
				   SELECT COALESCE(MAX(created_at), '0001-01-01 00:00:00')
					 FROM jobs
					WHERE type = $4
					      AND related_id=$5
			   );`, data.JobActive, data.JobDone, data.JobClientAfterChannelTopUp,
		data.JobClientPreChannelTopUp, chID)
	var qty int
	err := row.Scan(&qty)
	if err != nil {
		logger.Error(fmt.Sprintf("could not check for after topup channel job existance: %v", err))
		return err, false
	}
	return nil, qty == 0
}

func (m *Monitor) fetchSuggestedGasPrice() (*big.Int, error) {
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return m.suggestor.SuggestGasPrice(ctx)
}

func (m *Monitor) isToBeTerminated(logger log.Logger,
	ch *data.Channel, offer *data.Offering) (bool, error) {

	if ch.ReceiptBalance != 0 && ch.TotalDeposit != 0 &&
		ch.ReceiptBalance == ch.TotalDeposit {
		logger.Debug("channel reached its max. deposit")
		return true, nil
	}

	logger.Debug("channel max. deposit is not reached")

	reached, err := m.maxInactiveTimeReached(ch, offer)
	if err != nil {
		logger.Error(err.Error())
		// TODO: add error
		return false, ErrGetConsumedUnits
	}
	return reached, nil
}

func (m *Monitor) maxInactiveTimeReached(
	ch *data.Channel, offer *data.Offering) (bool, error) {
	query := fmt.Sprintf("SELECT COUNT(*), MAX(last_usage_time) FROM sessions WHERE sessions.channel = %s", m.db.Placeholder(1))
	var qty uint
	var lastUsageNullable pq.NullTime
	if err := m.db.QueryRow(
		query, ch.ID).Scan(&qty, &lastUsageNullable); err != nil {
		return false, err
	}
	lastUsage := lastUsageNullable.Time
	if qty == 0 {
		lastUsage = ch.PreparedAt
	}
	inactiveSeconds := uint64(time.Since(lastUsage).Seconds())
	return qty > 0 && inactiveSeconds > offer.MaxInactiveTimeSec, nil
}

func (m *Monitor) postCheque(channelID string, amount uint64) {
	logger := m.logger.Add("method", "posting cheque", "channel", channelID,
		"amount", amount)
	logger.Info("posting cheque")

	handleErr := func(err error) {
		select {
		case m.postChequeErrors <- err:
		default:
		}
	}

	var channel data.Channel
	if err := m.db.FindByPrimaryKeyTo(&channel, channelID); err != nil {
		logger.Error(err.Error())
		go handleErr(err)
		return
	}

	var client data.Account
	if err := m.db.FindOneTo(&client, "eth_addr", channel.Client); err != nil {
		logger.Error(err.Error())
		go handleErr(err)
		return
	}

	pscHex := data.HexFromBytes(common.HexToAddress(m.psc).Bytes())
	key, err := m.pw.GetKey(&client)
	if err != nil {
		logger.Error(err.Error())
		go handleErr(err)
		return
	}
	err = m.post(m.db, &channel, pscHex, key, amount,
		m.conf.RequestTLS, m.conf.RequestTimeout, m.pr)
	if err != nil {
		if err2, ok := err.(*srv.Error); ok {
			msg := fmt.Sprintf("%s (%d)", err2.Message, err2.Code)
			if err2.Code == pay.ErrCodeEqualBalance {
				logger.Debug(msg)
			} else {
				logger.Error(msg)
				go handleErr(err)
			}
			return
		}
		logger.Error(err.Error())
		go handleErr(err)
		return
	}

	logger.Info(fmt.Sprintf("sent payment channel: %s, amount: %v", channel, amount))
	res, err := m.db.Exec(`
		UPDATE channels
		   SET receipt_balance = $1
		 WHERE id = $2 AND receipt_balance < $1`, amount, channelID)
	if err != nil {
		logger.Error(err.Error())
		go handleErr(err)
		return
	}

	n, err := res.RowsAffected()
	if err != nil {
		if n != 0 {
			logger.Info("updated receipt balance")
		} else {
			logger.Warn("receipt balance isn't updated")
		}
	}
	go handleErr(err)
}
