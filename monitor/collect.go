package monitor

import (
	"context"
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
	freshBlocksKey      = "eth.event.freshblocks"
	blockLimitKey       = "eth.event.blocklimit"
	txMinedStatus       = "mined"
)

// Client defines typed wrappers for the Ethereum RPC API.
type Client interface {
	FilterLogs(ctx context.Context,
		q ethereum.FilterQuery) ([]ethtypes.Log, error)
	HeaderByNumber(ctx context.Context,
		number *big.Int) (*ethtypes.Header, error)
}

func hexesToHashes(hexes ...string) []common.Hash {
	hashes := make([]common.Hash, len(hexes))
	for i, hex := range hexes {
		hashes[i] = common.HexToHash(hex)
	}
	return hashes
}

var clientRelatedEvents = hexesToHashes(
	eth.EthDigestChannelToppedUp,
	eth.EthChannelCloseRequested,
	eth.EthOfferingEndpoint,
	eth.EthTokenApproval,
	eth.EthTokenTransfer,
)

var offeringRelatedEvents = hexesToHashes(
	eth.EthDigestChannelCreated,
	eth.EthCooperativeChannelClose,
	eth.EthUncooperativeChannelClose,
	eth.EthOfferingCreated,
	eth.EthOfferingDeleted,
	eth.EthOfferingPoppedUp,
)

// collect requests new logs and puts them into the database.
// timeout variable in seconds.
func (m *Monitor) collect(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx,
		time.Duration(m.cfg.CollectTimeout)*time.Second)
	defer cancel()

	firstBlock, lastBlock, err := m.getRangeOfInterest(ctx)
	if err != nil {
		m.errors <- err
		return
	}

	addresses, err := m.getAddressesInUse()
	if err != nil {
		m.errors <- err
		return
	}

	addressMap := make(map[common.Hash]bool)
	for _, a := range addresses {
		addressMap[a] = true
	}

	if firstBlock > lastBlock {
		m.logger.Add("firstBlock", firstBlock, "lastBlock",
			lastBlock).Debug("monitor has nothing to collect")
		return
	}
	m.logger.Add("firstBlock", firstBlock, "lastBlock",
		lastBlock).Debug("monitor is collecting logs")

	agentQ := ethereum.FilterQuery{
		Addresses: []common.Address{m.pscAddr, m.ptcAddr},
		FromBlock: new(big.Int).SetUint64(firstBlock),
		ToBlock:   new(big.Int).SetUint64(lastBlock),
		Topics:    [][]common.Hash{nil, addresses},
	}

	clientQ := agentQ
	clientQ.Topics = [][]common.Hash{clientRelatedEvents, nil, addresses}

	offeringQ := agentQ
	offeringQ.Topics = [][]common.Hash{offeringRelatedEvents}

	queries := []*ethereum.FilterQuery{&agentQ, &clientQ, &offeringQ}

	err = m.db.InTransaction(func(tx *reform.TX) error {
		for _, q := range queries {
			events, err := m.eth.FilterLogs(ctx, *q)
			if err != nil {
				m.logger.Error(err.Error())
				return ErrFetchLogs
			}

			for i := range events {
				e := &events[i]
				offeringRelated := q == &offeringQ
				forAgent := len(e.Topics) > 1 &&
					addressMap[e.Topics[1]]

				if e.Removed || offeringRelated && forAgent {
					continue
				}

				if err := m.collectEvent(tx, e); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		m.errors <- err
		return
	}

	m.setLastProcessedBlockNumber(lastBlock)
}

func (m *Monitor) collectEvent(tx *reform.TX, e *ethtypes.Log) error {
	el := &data.EthLog{
		ID:          util.NewUUID(),
		TxHash:      data.HexFromBytes(e.TxHash.Bytes()),
		TxStatus:    txMinedStatus,
		BlockNumber: e.BlockNumber,
		Addr:        data.HexFromBytes(e.Address.Bytes()),
		Data:        data.FromBytes(e.Data),
		Topics:      e.Topics,
	}
	if err := tx.Insert(el); err != nil {
		m.logger.Add("txHash", e.TxHash.String(),
			"blockNumber", e.BlockNumber, "contractAddress",
			e.Address.String()).Error(err.Error())
		return ErrInsertLogEvent
	}

	return nil
}

func (m *Monitor) getAddressesInUse() ([]common.Hash, error) {
	rows, err := m.db.Query(`SELECT eth_addr
		                         FROM accounts
		                        WHERE in_use`)
	if err != nil {
		m.logger.Error(err.Error())
		return nil, ErrGetActiveAccounts
	}
	defer rows.Close()

	var addresses []common.Hash
	for rows.Next() {
		var addrHex string
		if err := rows.Scan(&addrHex); err != nil {
			return nil, fmt.Errorf("failed to scan rows: %v", err)
		}
		addresses = append(addresses, common.HexToHash(addrHex))
	}
	if err := rows.Err(); err != nil {
		m.logger.Error(err.Error())
		return nil, ErrTraverseAddresses
	}
	return addresses, nil
}

// getRangeOfInterest returns the range of block numbers
// that need to be scanned for new logs. It respects
// the min confirmations setting.
func (m *Monitor) getRangeOfInterest(
	ctx context.Context) (first, last uint64, err error) {
	unreliableNum, err := data.GetUint64Setting(m.db, minConfirmationsKey)
	if err != nil {
		m.logger.Add("setting", minConfirmationsKey).Error(err.Error())
		return 0, 0, err
	}

	freshNum, err := data.GetUint64Setting(m.db, freshBlocksKey)
	if err != nil {
		m.logger.Add("setting", freshBlocksKey).Error(err.Error())
		return 0, 0, err
	}

	first, err = m.getLastProcessedBlockNumber()
	if err != nil {
		return 0, 0, err
	}

	first = first + 1

	latestBlock, err := m.getLatestBlockNumber(ctx)
	if err != nil {
		return 0, 0, err
	}

	last = safeSub(latestBlock, unreliableNum)

	if freshNum != 0 {
		first = max(first, safeSub(last, freshNum))
	}

	limitNum, err := data.GetUint64Setting(m.db, blockLimitKey)
	if err != nil {
		m.logger.Add("setting", blockLimitKey).Warn(err.Error())
		return first, last, nil
	}

	if limitNum != 0 && last > first && (last-first) > limitNum {
		last = first + limitNum
	}

	return first, last, nil
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

func (m *Monitor) getLastProcessedBlockNumber() (uint64, error) {
	if m.lastProcessedBlock == 0 {
		row := m.db.QueryRow(`SELECT MAX(block_number)
					      FROM eth_logs`)
		var v *uint64

		if err := row.Scan(&v); err != nil {
			m.logger.Error(err.Error())
			return 0, ErrScanRows
		}
		if v != nil {
			m.mtx.Lock()
			m.lastProcessedBlock = *v
			m.mtx.Unlock()
		}
	}

	return m.lastProcessedBlock, nil
}

func (m *Monitor) setLastProcessedBlockNumber(number uint64) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.lastProcessedBlock = number
}

func (m *Monitor) getLatestBlockNumber(ctx context.Context) (uint64, error) {
	ctx, cancel := context.WithTimeout(ctx,
		time.Duration(m.callTimeout)*time.Second)
	defer cancel()

	header, err := m.eth.HeaderByNumber(ctx, nil)
	if err != nil {
		m.logger.Error(err.Error())
		return 0, ErrGetHeaderByNumber
	}

	return header.Number.Uint64(), err
}
