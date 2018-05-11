package monitor

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/util"
)

const (
	minConfirmationsKey = "eth.min.confirmations"
)

// Monitor implements blockchain monitor which fetches logs from the blockchain
// and creates jobs accordingly.
type Monitor struct {
	logger  *util.Logger
	db      *reform.DB
	eth     *ethclient.Client
	pscAddr common.Address

	lastProcessedBlock uint64

	cancel context.CancelFunc
}

// NewMonitor creates a Monitor with specified settings.
func NewMonitor(
	logger *util.Logger,
	db *reform.DB,
	eth *ethclient.Client,
	pscAddr common.Address) *Monitor {
	return &Monitor{
		logger: logger,
		db: db,
		eth: eth,
		pscAddr: pscAddr,
	}
}

// Start starts the monitor. It will continue collecting logs and scheduling
// jobs until it is stopped with Stop.
func (m *Monitor) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	d := 10 * time.Second // FIXME: hardcoded duration
	go m.repeatEvery(ctx, d, "collect", func() {m.collect(ctx)})
	// go m.repeatEvery(ctx, d, "schedule", func() {m.schedule()})

	m.logger.Debug("monitor started")
	return nil
}

// Stop makes the monitor stop.
func (m *Monitor) Stop() error {
	m.cancel()

	m.logger.Debug("monitor stopped")
	return nil
}

// repeatEvery calls a given action function repeatedly with period d. It
// recovers and continues repeating in case of panic. To stop the loop,
// cancel the context.
func (m *Monitor) repeatEvery(ctx context.Context, d time.Duration, name string, action func()) {
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error("recovered in monitor's %s loop: %v", name, r)
			go m.repeatEvery(ctx, d, name, action)
		}
	}()

	ticker := time.NewTicker(d)
	defer ticker.Stop()

	for {
		select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				action()
		}
	}
}

var clientRelatedEvents = []common.Hash{
	common.HexToHash(eth.EthDigestChannelCreated),
	common.HexToHash(eth.EthDigestChannelToppedUp),
	common.HexToHash(eth.EthChannelCloseRequested),
	common.HexToHash(eth.EthOfferingEndpoint),
	common.HexToHash(eth.EthCooperativeChannelClose),
	common.HexToHash(eth.EthUncooperativeChannelClose),
}

// collect requests new logs and puts them into the database.
func (m *Monitor) collect(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 5 * time.Second) // FIXME: hardcoded duration
	defer cancel()

	pscAddr := []common.Address{m.pscAddr}
	fromBlock, toBlock := m.blocksOfInterest(ctx)
	m.logger.Debug(
		"monitor is collecting logs from blocks %d to %d",
		fromBlock,
		toBlock,
	)
	if fromBlock > toBlock {
		m.logger.Debug("monitor has nothing to collect")
		return
	}

	addresses := m.getAddressesInUse()

	agentQ := ethereum.FilterQuery{
		Addresses: pscAddr,
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Topics:    [][]common.Hash{nil, addresses},
	}

	clientQ := agentQ
	clientQ.Topics = [][]common.Hash{clientRelatedEvents, nil, addresses}

	queries := []*ethereum.FilterQuery{&agentQ, &clientQ}

	err := m.db.InTransaction(func(tx *reform.TX) error {
		for _, q := range queries {
			events, err := m.eth.FilterLogs(ctx, *q)
			if err != nil {
				return fmt.Errorf("could not fetch logs over rpc: %v", err)
			}

			for i := range events {
				m.collectEvent(tx, &events[i])
			}
		}

		return nil
	})
	if err != nil {
		panic(fmt.Errorf("log collecting failed: %v", err))
	}

	m.setLastProcessedBlockNumber(toBlock)
}

func (m *Monitor) collectEvent(tx *reform.TX, e *ethtypes.Log) {
	if e.Removed {
		return
	}

	var topics data.LogTopics
	for _, hash := range e.Topics {
		topics = append(topics, hash.Hex())
	}

	le := &data.LogEntry{
		ID: util.NewUUID(),
		TxHash: data.FromBytes(e.TxHash.Bytes()),
		TxStatus: "mined", // FIXME: is this field needed at all?
		BlockNumber: e.BlockNumber,
		Addr: data.FromBytes(e.Address.Bytes()),
		Data: data.FromBytes(e.Data),
		Topics: topics,
	}

	if err := tx.Insert(le); err != nil {
		panic(fmt.Errorf("failed to insert a log event into db: %v", err))
	}
}

func (m *Monitor) getAddressesInUse() []common.Hash {
	rows, err := m.db.Query("select eth_addr from accounts where in_use")
	if err != nil {
		panic(fmt.Errorf("failed to query active accounts from db: %v", err))
	}
	defer rows.Close()

	var addresses []common.Hash
	for rows.Next() {
		var b64 string
		rows.Scan(&b64)
		addrBytes, err := data.ToBytes(b64)
		if err != nil {
			panic(fmt.Errorf("failed to decode eth address from base64: %v", err))
		}
		addresses = append(addresses, common.BytesToHash(addrBytes))
	}
	if err := rows.Err(); err != nil {
		panic(fmt.Errorf("failed to traverse the selected eth addresses: %v", err))
	}
	return addresses
}

// blocksOfInterest returns the range of block numbers that need to be scanned
// for new logs. It respects the min confirmations setting.
func (m *Monitor) blocksOfInterest(ctx context.Context) (from, to uint64) {
	minConfirmations := m.getMinConfirmations()

	from = m.getLastProcessedBlockNumber() + 1
	to = m.getLatestBlockNumber(ctx)
	if minConfirmations < to {
		to -= minConfirmations
	} else {
		to = 0
	}

	return from, to
}

func (m *Monitor) getLastProcessedBlockNumber() uint64 {
	if m.lastProcessedBlock == 0 {
		row := m.db.QueryRow("select max(block_number) from eth_logs")
		var v *uint64
		err := row.Scan(&v)
		if err != nil && err != sql.ErrNoRows {
			panic(err)
		}
		if v != nil {
			m.lastProcessedBlock = *v
		}
	}

	return m.lastProcessedBlock
}

func (m *Monitor) setLastProcessedBlockNumber(number uint64) {
	m.lastProcessedBlock = number
}

func (m *Monitor) getMinConfirmations() uint64 {
	var setting data.Setting
	err := m.db.FindByPrimaryKeyTo(&setting, minConfirmationsKey)
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
		m.logger.Error("failed to parse %s setting: %v", minConfirmationsKey, err)
		return 0
	}

	return value
}

func (m *Monitor) getLatestBlockNumber(ctx context.Context) uint64 {
	ctx, cancel := context.WithTimeout(ctx, 5 * time.Second)
	defer cancel()

	header, err := m.eth.HeaderByNumber(ctx, nil)
	if err != nil {
		panic(err)
	}

	return header.Number.Uint64()
}
