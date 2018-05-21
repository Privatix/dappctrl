package monitor

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/util"
)

const (
	minConfirmationsKey = "eth.min.confirmations"
	freshOfferingsKey   = "eth.event.freshofferings"
)

type Client interface {
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]ethtypes.Log, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*ethtypes.Header, error)
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
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second) // FIXME: hardcoded duration
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

func (m *Monitor) collectEvent(tx *reform.TX, e *ethtypes.Log) {
	var topics data.LogTopics
	for _, hash := range e.Topics {
		topics = append(topics, hash.Hex())
	}

	le := &data.LogEntry{
		ID:          util.NewUUID(),
		TxHash:      data.FromBytes(e.TxHash.Bytes()),
		TxStatus:    "mined", // FIXME: is this field needed at all?
		BlockNumber: e.BlockNumber,
		Addr:        data.FromBytes(e.Address.Bytes()),
		Data:        data.FromBytes(e.Data),
		Topics:      topics,
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

func (m *Monitor) getLatestBlockNumber(ctx context.Context) uint64 {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second) // FIXME: hardcoded timeout
	defer cancel()

	header, err := m.eth.HeaderByNumber(ctx, nil)
	if err != nil {
		panic(err)
	}

	return header.Number.Uint64()
}
