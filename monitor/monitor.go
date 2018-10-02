package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/util/log"
)

const (
	collectName  = "collect"
	scheduleName = "schedule"
)

type heartbeat struct {
	collect  chan struct{}
	schedule chan struct{}
}

// Config for blockchain monitor.
type Config struct {
	CollectPause    int64 // pause between collect iterations
	SchedulePause   int64 // pause between schedule iterations
	CollectTimeout  uint64
	ScheduleTimeout uint64
}

// Queue is a job processing queue.
type Queue interface {
	Add(j *data.Job) error
}

// Monitor implements blockchain monitor which fetches logs from the blockchain
// and creates jobs accordingly.
type Monitor struct {
	cfg     *Config
	ctx     context.Context
	ethCfg  *eth.Config
	logger  log.Logger
	db      *reform.DB
	queue   Queue
	eth     Client
	pscAddr common.Address
	pscABI  abi.ABI
	ptcAddr common.Address

	mtx                sync.Mutex
	lastProcessedBlock uint64

	cancel               context.CancelFunc
	errors               chan error
	dappRole             string
	tickers              []*time.Ticker
	callTimeout          uint64
	closeIdleConnections func()
}

// NewConfig creates a default blockchain monitor configuration.
func NewConfig() *Config {
	return &Config{
		CollectPause:    6,
		SchedulePause:   6,
		CollectTimeout:  60,
		ScheduleTimeout: 5,
	}
}

// NewMonitor creates a Monitor with specified settings.
func NewMonitor(cfg *Config, logger log.Logger, db *reform.DB,
	queue Queue, ethConfig *eth.Config, pscAddr common.Address,
	ptcAddr common.Address, ethClient Client, dappRole string,
	closeIdleConnections func()) (*Monitor, error) {
	if logger == nil || db == nil || queue == nil ||
		!common.IsHexAddress(pscAddr.String()) {
		return nil, ErrInput
	}

	pscABI, err := mustParseABI(contract.PrivatixServiceContractABI)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrParseABI
	}

	return &Monitor{
		cfg:                  cfg,
		ethCfg:               ethConfig,
		eth:                  ethClient,
		logger:               logger.Add("type", "monitor.Monitor"),
		db:                   db,
		queue:                queue,
		pscAddr:              pscAddr,
		pscABI:               pscABI,
		ptcAddr:              ptcAddr,
		mtx:                  sync.Mutex{},
		errors:               make(chan error),
		dappRole:             dappRole,
		callTimeout:          ethConfig.Timeout,
		closeIdleConnections: closeIdleConnections,
	}, nil
}

func (m *Monitor) errorProcessing(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-m.errors:
			// errors associated with blockchain
			if (err == ErrGetHeaderByNumber ||
				err == ErrFetchLogs) &&
				m.closeIdleConnections != nil {
				m.closeIdleConnections()
			}
		}
	}
}

func (m *Monitor) start(ctx context.Context, collectTicker,
	scheduleTicker <-chan time.Time) *heartbeat {
	collectSignalChan := make(chan struct{})
	scheduleSignalChan := make(chan struct{})

	go m.errorProcessing(ctx)
	go m.repeatEvery(ctx, collectTicker, collectName,
		func() { m.collect(ctx) }, collectSignalChan)
	go m.repeatEvery(ctx, scheduleTicker, scheduleName,
		func() { m.schedule(ctx) }, scheduleSignalChan)

	m.logger.Debug("blockchain monitor started")

	return &heartbeat{
		collect:  collectSignalChan,
		schedule: scheduleSignalChan,
	}
}

// Start starts the monitor. It will continue collecting logs and scheduling
// jobs until it is stopped with Stop.
func (m *Monitor) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	collectTicker := time.NewTicker(
		time.Duration(m.cfg.CollectPause) * time.Second)
	scheduleTicker := time.NewTicker(
		time.Duration(m.cfg.SchedulePause) * time.Second)
	m.tickers = append(m.tickers, collectTicker, scheduleTicker)
	m.start(ctx, collectTicker.C, scheduleTicker.C)
	return nil
}

// Stop makes the monitor stop.
func (m *Monitor) Stop() error {
	m.cancel()
	for _, t := range m.tickers {
		t.Stop()
	}

	m.logger.Debug("blockchain monitor stopped")
	return nil
}

// repeatEvery calls a given action function repeatedly every time a read on
// ticker channel succeeds. To stop the loop, cancel the context.
func (m *Monitor) repeatEvery(ctx context.Context, ticker <-chan time.Time,
	name string, action func(), signalChan chan struct{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
			action()
			select {
			case signalChan <- struct{}{}:
			default:
			}
		}
	}
}
