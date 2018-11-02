package monitor

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/util/log"
)

// Client defines typed wrappers for the Ethereum RPC API.
type Client interface {
	FilterLogs(ctx context.Context,
		q ethereum.FilterQuery) ([]ethtypes.Log, error)
	HeaderByNumber(ctx context.Context,
		number *big.Int) (*ethtypes.Header, error)
}

// Config is a monitor configuration.
type Config struct {
	EthCallTimeout uint // In seconds.
	QueryPause     uint // In seconds.
}

// NewConfig creates a default blockchain monitor configuration.
func NewConfig() *Config {
	return &Config{
		QueryPause:     6,
		EthCallTimeout: 60,
	}
}

// Queue is a job queue.
type Queue interface {
	Add(*reform.TX, *data.Job) error
}

// Monitor is a blockchain monitor.
type Monitor struct {
	db     *reform.DB
	eth    Client
	logger log.Logger
	queue  Queue

	ethCallTimeout      time.Duration
	pscABI              abi.ABI
	pscAddr             common.Address
	stopMonitor         func()
	queryPause          time.Duration
	getFilterLogQueries queriesBuilderFunc
	jobsProducers       JobsProducers
}

// NewMonitor creates new blockchain monitor.
func NewMonitor(conf *Config, c Client, db *reform.DB, l log.Logger, psc,
	ptc common.Address, role string, q Queue) (*Monitor, error) {
	abiJSON := contract.PrivatixServiceContractABI
	pscABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		l.Error(err.Error())
		return nil, ErrFailedToParseABI
	}
	ethCallTimeout := time.Duration(conf.EthCallTimeout) * time.Second
	queryPause := time.Duration(conf.QueryPause) * time.Second
	m := &Monitor{
		db:             db,
		eth:            c,
		logger:         l.Add("type", "monitor.Monitor"),
		queue:          q,
		ethCallTimeout: ethCallTimeout,
		pscABI:         pscABI,
		pscAddr:        psc,
		queryPause:     queryPause,
	}

	f := m.clientQueries
	m.jobsProducers = m.clientJobsProducers()
	if role == data.RoleAgent {
		f = m.agentQueries
		m.jobsProducers = m.agentJobsProducers()
	}
	m.getFilterLogQueries = func(from, to uint64) ([]ethereum.FilterQuery, error) {
		return f(from, to, psc, ptc)
	}

	return m, nil
}

// Start starts scanning blockchain for events.
func (m *Monitor) Start() {
	logger := m.logger.Add("method", "Start")
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(m.queryPause)
	m.stopMonitor = func() {
		ticker.Stop()
		cancel()
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := m.queryLogsAndCreateJobs(
					m.getFilterLogQueries, m.jobsProducers)
				if err != nil {
					logger.Error(err.Error())
				}
			}
		}
	}()
}

// Stop stops the monitor.
func (m *Monitor) Stop() {
	m.stopMonitor()
}
