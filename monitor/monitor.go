package monitor

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/util"
)

const (
	minConfirmationsKey = "eth.min.confirmations"
	freshOfferingsKey = "eth.event.freshofferings"
)

type Client interface {
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]ethtypes.Log, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*ethtypes.Header, error)
}

// Monitor implements blockchain monitor which fetches logs from the blockchain
// and creates jobs accordingly.
type Monitor struct {
	logger  *util.Logger
	db      *reform.DB
	eth     Client
	pscAddr common.Address

	lastProcessedBlock uint64

	cancel context.CancelFunc
	tickers []*time.Ticker
}

// NewMonitor creates a Monitor with specified settings.
func NewMonitor(
	logger *util.Logger,
	db *reform.DB,
	eth Client,
	pscAddr common.Address) *Monitor {
	return &Monitor{
		logger: logger,
		db: db,
		eth: eth,
		pscAddr: pscAddr,
	}
}

func (m *Monitor) start(ctx context.Context, collectTicker, scheduleTicker <-chan time.Time) error {
	go m.repeatEvery(ctx, collectTicker, "collect", func() {m.collect(ctx)})
	go m.repeatEvery(ctx, scheduleTicker, "schedule", func() {m.schedule(ctx)})

	m.logger.Debug("monitor started")
	return nil
}

// Start starts the monitor. It will continue collecting logs and scheduling
// jobs until it is stopped with Stop.
func (m *Monitor) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	collectTicker := time.NewTicker(10 * time.Second) // FIXME: hardcoded period
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

var clientRelatedEvents = []common.Hash{
	common.HexToHash(eth.EthDigestChannelCreated),
	common.HexToHash(eth.EthDigestChannelToppedUp),
	common.HexToHash(eth.EthChannelCloseRequested),
	common.HexToHash(eth.EthOfferingEndpoint),
	common.HexToHash(eth.EthCooperativeChannelClose),
	common.HexToHash(eth.EthUncooperativeChannelClose),
}

var offeringRelatedEvents = []common.Hash{
	common.HexToHash(eth.EthOfferingCreated),
	common.HexToHash(eth.EthOfferingDeleted),
	common.HexToHash(eth.EthOfferingPoppedUp),
}

/*
Monitor performs Log collecting periodically.

Several most recent blocks on the blockchain are considered "unreliable" (the
relevant setting is "eth.min.confirmations").

Let A = last processed block number
    Z = most recent block number on the blockchain
    C = the min confirmations setting
    F = the fresh offerings setting

Thus the range of interest for agent and client logs Ri = [A + 1, Z - C],
and for offering it is:

    if F > 0 Ro = Ri ∩ [Z - C - F, +inf)
    else Ro = Ri

These are the rules for filtering logs on the blockchain:

1. Events for agent
   - From: A + 1
   - To:   Z - C
   - Topics[0]: any
   - Topics[1]: one of accounts with in_use = true

2. Events for client
   - From: A + 1
   - To:   Z - C
   - Topics[0]: one of these hashes
     - LogChannelCreated
     - LogChannelToppedUp
     - LogChannelCloseRequested
     - LogCooperativeChannelClose
     - LogUnCooperativeChannelClose
   - Topics[2]: one of the accounts with in_use = true

3. Offering events
   - From: max(A + 1, Z - C - F) if F > 0
   - From: A + 1 if F == 0
   - To:   Z - C
   - Topics[0]: one of these hashes
     - LogOfferingCreated
     - LogOfferingDeleted
     - LogOfferingPopedUp
   - Topics[1]: not one of the accounts with in_use = true
   - Topics[2]: one of the accounts with in_use = true
*/

// collect requests new logs and puts them into the database.
func (m *Monitor) collect(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 5 * time.Second) // FIXME: hardcoded duration
	defer cancel()

	pscAddr := []common.Address{m.pscAddr}
	firstBlock, freshBlock, lastBlock := m.getRangeOfInterest(ctx)
	addresses := m.getAddressesInUse()
	addressMap := make(map[common.Hash]bool)
	for _, a := range addresses {
		addressMap[a] = true
	}

	if firstBlock > lastBlock {
		m.logger.Debug("monitor has nothing to collect")
		return
	}
	m.logger.Debug(
		"monitor is collecting logs from blocks %d to %d",
		firstBlock,
		lastBlock,
	)

	agentQ := ethereum.FilterQuery{
		Addresses: pscAddr,
		FromBlock: new(big.Int).SetUint64(firstBlock),
		ToBlock:   new(big.Int).SetUint64(lastBlock),
		Topics:    [][]common.Hash{nil, addresses},
	}

	clientQ := agentQ
	clientQ.Topics = [][]common.Hash{clientRelatedEvents, nil, addresses}

	offeringQ := agentQ
	offeringQ.FromBlock = new(big.Int).SetUint64(freshBlock)
	offeringQ.Topics = [][]common.Hash{offeringRelatedEvents}

	queries := []*ethereum.FilterQuery{&agentQ, &clientQ, &offeringQ}

	err := m.db.InTransaction(func(tx *reform.TX) error {
		for _, q := range queries {
			events, err := m.eth.FilterLogs(ctx, *q)
			if err != nil {
				return fmt.Errorf("could not fetch logs over rpc: %v", err)
			}

			for i := range events {
				e := &events[i]
				offeringRelated := q == &offeringQ
				forAgent := len(e.Topics) > 1 && addressMap[e.Topics[1]]

				if e.Removed || offeringRelated && forAgent {
					continue
				}

				m.collectEvent(tx, e)
			}
		}

		return nil
	})
	if err != nil {
		panic(fmt.Errorf("log collecting failed: %v", err))
	}

	m.setLastProcessedBlockNumber(lastBlock)
}

// schedule creates a job for each unprocessed log event in the database.
func (m *Monitor) schedule(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 5 * time.Second) // FIXME: hardcoded duration
	defer cancel()

	// TODO: Move this logic into a database view? The query is just supposed to
	// append two boolean columns calculated based on topics: whether the
	// event is for agent, and the same for client.
	//
	// eth_logs.topics is a json array with '0xdeadbeef' encoding of addresses,
	// whereas accounts.eth_addr is a base64 encoding of raw bytes of addresses.
	// The encode-decode-substr is there to convert from one to another.
	// coalesce() converts null into false for the case when topics->>n does not exist.
	topicInAccExpr := `
		coalesce(
			encode(decode(substr(topics->>%d, 3), 'hex'), 'base64')
			in (select eth_addr from accounts where in_use),
			false
		)
	`
	columns := m.db.QualifiedColumns(data.LogEntryTable)
	columns = append(columns, fmt.Sprintf(topicInAccExpr, 1)) // topic[1] (agent) in active accounts
	columns = append(columns, fmt.Sprintf(topicInAccExpr, 2)) // topic[2] (client) in active accounts

	query := fmt.Sprintf(
		"select %s from eth_logs where job IS NULL",
		strings.Join(columns, ","),
	)
	rows, err := m.db.Query(query)
	if err != nil {
		panic(fmt.Errorf("failed to select log entries: %v", err))
	}

	for rows.Next() {
		var e data.LogEntry
		var forAgent, forClient bool
		pointers := append(e.Pointers(), &forAgent, &forClient)
		if err := rows.Scan(pointers...); err != nil {
			panic(fmt.Errorf("failed to scan the selected log entries: %v", err))
		}

		if forAgent {
			m.scheduleForAgent(&e)
		} else if forClient {
			m.scheduleForClient(&e)
		} else {
			m.ignoreEvent(&e)
		}
	}
	if err := rows.Err(); err != nil {
		panic(fmt.Errorf("failed to fetch the next selected log entry: %v", err))
	}
}

// schedule creates an agent job for a given event.
func (m *Monitor) scheduleForAgent(e *data.LogEntry) {
}

// schedule creates a client job for a given event.
func (m *Monitor) scheduleForClient(e *data.LogEntry) {
}

func (m *Monitor) ignoreEvent(e *data.LogEntry) {
	zero := "00000000-0000-0000-0000-000000000000"
	e.JobID = &zero
	if err := m.db.Save(e); err != nil {
		panic(fmt.Errorf("failed to ignore event: %v", err))
	}
}

func (m *Monitor) collectEvent(tx *reform.TX, e *ethtypes.Log) {
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

// getRangeOfInterest returns the range of block numbers that need to be scanned
// for new logs. It respects the min confirmations setting.
func (m *Monitor) getRangeOfInterest(ctx context.Context) (first, fresh, last uint64) {
	unreliableNum := m.getUint64Setting(minConfirmationsKey)
	freshNum := m.getUint64Setting(freshOfferingsKey)

	first = m.getLastProcessedBlockNumber() + 1
	last = safeSub(m.getLatestBlockNumber(ctx), unreliableNum)

	if freshNum == 0 {
		fresh = first
	} else {
		fresh = max(first, safeSub(last, freshNum))
	}

	return first, fresh, last
}

func safeSub(a, b uint64) uint64 {
	if a > b {
		return a - b
	}
	return 0
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
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

func (m *Monitor) getLatestBlockNumber(ctx context.Context) uint64 {
	ctx, cancel := context.WithTimeout(ctx, 5 * time.Second) // FIXME: hardcoded timeout
	defer cancel()

	header, err := m.eth.HeaderByNumber(ctx, nil)
	if err != nil {
		panic(err)
	}

	return header.Number.Uint64()
}
