package monitor

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

type Queue interface {
	Add(j *data.Job) error
}

// Monitor implements blockchain monitor which fetches logs from the blockchain
// and creates jobs accordingly.
type Monitor struct {
	logger  *util.Logger
	db      *reform.DB
	queue   Queue
	eth     Client
	pscAddr common.Address

	lastProcessedBlock uint64

	cancel  context.CancelFunc
	tickers []*time.Ticker
}

// NewMonitor creates a Monitor with specified settings.
func NewMonitor(
	logger *util.Logger,
	db *reform.DB,
	queue Queue,
	eth Client,
	pscAddr common.Address) *Monitor {
	return &Monitor{
		logger:  logger,
		db:      db,
		queue:   queue,
		eth:     eth,
		pscAddr: pscAddr,
	}
}

func (m *Monitor) start(ctx context.Context, collectTicker, scheduleTicker <-chan time.Time) error {
	go m.repeatEvery(ctx, collectTicker, "collect", func() { m.collect(ctx) })
	go m.repeatEvery(ctx, scheduleTicker, "schedule", func() { m.schedule(ctx) })

	m.logger.Debug("monitor started")
	return nil
}

// Start starts the monitor. It will continue collecting logs and scheduling
// jobs until it is stopped with Stop.
func (m *Monitor) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	collectTicker := time.NewTicker(10 * time.Second)  // FIXME: hardcoded period
	scheduleTicker := time.NewTicker(10 * time.Second) // FIXME: hardcoded period
	m.tickers = append(m.tickers, collectTicker, scheduleTicker)
	return m.start(ctx, collectTicker.C, scheduleTicker.C)
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
// ticker channel succeeds. It recovers and continues repeating in case of
// panic. To stop the loop, cancel the context.
func (m *Monitor) repeatEvery(ctx context.Context, ticker <-chan time.Time, name string, action func()) {
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error("recovered in monitor's %s loop: %v", name, r)
			go m.repeatEvery(ctx, ticker, name, action)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
			action()
		}
	}
}

func (m *Monitor) getUint64Setting(key string) uint64 {
	var setting data.Setting
	err := m.db.FindByPrimaryKeyTo(&setting, key)
	switch err {
	case nil:
		break
	case sql.ErrNoRows:
		return 0
	default:
		panic(err)
	}

	value, err := strconv.ParseUint(setting.Value, 10, 64)
	if err != nil {
		m.logger.Error("failed to parse %s setting: %v", key, err)
		return 0
	}

	return value
}
