package monitor

import (
	"context"
	"fmt"
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
	EthCallTimeout uint   // In milliseconds.
	InitialBlocks  uint64 // In Ethereum blocks.
	QueryPause     uint   // In milliseconds.
}

// NewConfig creates a default blockchain monitor configuration.
func NewConfig() *Config {
	return &Config{
		EthCallTimeout: 60000,
		InitialBlocks:  5760, // Is equivalent to 24 hours.
		QueryPause:     6000,
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
	ptcABI              abi.ABI
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
	abiJSON2 := contract.PrivatixTokenContractABI
	ptcABI, err := abi.JSON(strings.NewReader(abiJSON2))
	if err != nil {
		l.Error(err.Error())
		return nil, ErrFailedToParseABI
	}
	ethCallTimeout := time.Duration(conf.EthCallTimeout) * time.Millisecond
	queryPause := time.Duration(conf.QueryPause) * time.Millisecond
	m := &Monitor{
		db:             db,
		eth:            c,
		logger:         l.Add("type", "monitor.Monitor"),
		queue:          q,
		ethCallTimeout: ethCallTimeout,
		pscABI:         pscABI,
		ptcABI:         ptcABI,
		pscAddr:        psc,
		queryPause:     queryPause,
	}

	m.initLastProcessedBlock(role, conf.InitialBlocks)

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

// initLastProcessedBlock calculates last processed block. If user role is
// client and value of "eth.event.lastProcessedBlock" setting is 0, then value
// of "eth.event.lastProcessedBlock" setting is equal the difference between
// the last Ethereum block and a InitialBlocks value. If InitialBlocks
// value is 0, then this parameter is ignored.
func (m *Monitor) initLastProcessedBlock(role string, initialBlocks uint64) {
	logger := m.logger.Add("method", "initLastProcessedBlock",
		"role", role, "initialBlocks", initialBlocks)

	block, err := m.getLastProcessedBlockNumber()
	if block > 0 || err != nil || role != data.RoleClient ||
		initialBlocks == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(),
		m.ethCallTimeout)
	defer cancel()

	lastBlock, err := m.eth.HeaderByNumber(ctx, nil)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	lastEthBlockNum := lastBlock.Number.Uint64()
	logger = logger.Add("lastEthBlockNum", lastEthBlockNum)

	if initialBlocks > lastEthBlockNum {
		logger.Warn("initialBlocks value is very big")
		return
	}

	lastProcessedBlock := lastEthBlockNum - initialBlocks

	_, err = m.db.Exec(`UPDATE settings SET value=$1 WHERE key=$2`,
		lastProcessedBlock, data.SettingLastProcessedBlock)
	if err != nil {
		logger.Error(err.Error())
	}
	logger.Debug(fmt.Sprintf("last processed block: %d",
		lastProcessedBlock))
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
					logger.Warn(err.Error())
				}
			}
		}
	}()
}

// Stop stops the monitor.
func (m *Monitor) Stop() {
	m.stopMonitor()
}
