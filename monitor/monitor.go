package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/util"
)

const (
	collectName  = "collect"
	scheduleName = "schedule"
)

// Blockchain monitor errors.
var (
	ErrInput = fmt.Errorf("one or more input parameters is wrong")
)

// Config for blockchain monitor.
type Config struct {
	CollectPause  int64 // pause between collect iterations
	SchedulePause int64 // pause between schedule iterations
	Timeout       int64 // maximum time of one operation
}

// Queue is a job processing queue.
type Queue interface {
	Add(j *data.Job) error
}

// Monitor implements blockchain monitor which fetches logs from the blockchain
// and creates jobs accordingly.
type Monitor struct {
	cfg     *Config
	logger  *util.Logger
	db      *reform.DB
	queue   Queue
	eth     Client
	pscAddr common.Address
	pscABI  abi.ABI
	ptcAddr common.Address

	mu                 sync.Mutex
	lastProcessedBlock uint64

	cancel  context.CancelFunc
	errors  chan error
	tickers []*time.Ticker
}

// NewConfig creates a default blockchain monitor configuration.
func NewConfig() *Config {
	return &Config{
		CollectPause:  6,
		SchedulePause: 6,
		Timeout:       5,
	}
}

// NewMonitor creates a Monitor with specified settings.
func NewMonitor(cfg *Config, logger *util.Logger, db *reform.DB,
	queue Queue, eth Client, pscAddr common.Address,
	ptcAddr common.Address) (*Monitor, error) {
	if logger == nil || db == nil || queue == nil || eth == nil ||
		cfg.CollectPause <= 0 || cfg.SchedulePause <= 0 ||
		cfg.Timeout <= 0 || !common.IsHexAddress(pscAddr.String()) {
		return nil, ErrInput
	}

	pscABI, err := mustParseABI(contract.PrivatixServiceContractABI)
	if err != nil {
		return nil, err
	}

	return &Monitor{
		cfg:     cfg,
		logger:  logger,
		db:      db,
		queue:   queue,
		eth:     eth,
		pscAddr: pscAddr,
		pscABI:  pscABI,
		ptcAddr: ptcAddr,
		mu:      sync.Mutex{},
		errors:  make(chan error),
	}, nil
}

func (m *Monitor) start(ctx context.Context, timeout int64, collectTicker,
	scheduleTicker <-chan time.Time, errCh chan error) {
	go m.repeatEvery(ctx, collectTicker, errCh, collectName,
		func() { m.collect(ctx, timeout, errCh) })
	go m.repeatEvery(ctx, scheduleTicker, errCh, scheduleName,
		func() { m.schedule(ctx, timeout, errCh) })

	m.logger.Debug("monitor started")
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
	m.start(ctx, m.cfg.Timeout, collectTicker.C,
		scheduleTicker.C, m.errors)
	return nil
}

// Stop makes the monitor stop.
func (m *Monitor) Stop() error {
	m.cancel()
	for _, t := range m.tickers {
		t.Stop()
	}

	m.logger.Debug("monitor stopped")
	return nil
}

// repeatEvery calls a given action function repeatedly every time a read on
// ticker channel succeeds. To stop the loop, cancel the context.
func (m *Monitor) repeatEvery(ctx context.Context, ticker <-chan time.Time,
	errCh chan error, name string, action func()) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
			action()
		case err := <-errCh:
			m.logger.Error("blockchain monitor: %s", err)
		}
	}
}
