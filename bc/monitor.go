package bc

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
)

// Client defines typed wrappers for the Ethereum RPC API.
type Client interface {
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
	HeaderByNumber(ctx context.Context,
		number *big.Int) (*types.Header, error)
}

// Queue is a job queue.
type Queue interface {
	Add(*reform.TX, *data.Job) error
}

// // QueriesMaker given the latest ethereum block returns filter ethereum logs queries.
// type QueriesMaker func(db *reform.Querier, latestBlock uint64) ([]ethereum.FilterQuery, error)

// // JobMaker returns new job to create for a given ethereum log.
// // DB querier used to appropriately fill job fields (releated_id etc).
// type JobMaker func(*reform.Querier, []data.JobEthLog) ([]data.Job, error)

// Monitor continuously scans blockchain for new ethereum logs and creates jobs for them.
//
// Each round of monitoring consists of adding jobs for ethereum logs filtered by
// filter queries. Each round start by NextRound and DoneRound.
type Monitor struct {
	// RequestTimeout is a timeout for requeting anything using Client.
	RequestTimeout time.Duration
	// Queue is a job queue.
	Queue Queue
	// NextRound returns queries to get ethereum logs at the next round of monitoring.
	NextRound func(latestBlock uint64) ([]ethereum.FilterQuery, func(*reform.TX) error, error)
	// JobsForLog return jobs to create for given ethereum log.
	JobsForLog func(*data.JobEthLog) ([]data.Job, error)
	// RoundsInterval is an interval at which monitoring rounds are run.
	RoundsInterval time.Duration

	client  Client
	db      *reform.DB
	logger  log.Logger
	stopMon func()
}

// Start starts monitoring rounds.
func (m *Monitor) Start() {
	logger := m.logger.Add("method", "Start")

	ctx, cancel := context.WithCancel(context.Background())
	m.stopMon = cancel

	tick := time.NewTicker(m.RoundsInterval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if err := m.Round(); err != nil {
				logger.Error(err.Error())
			}
		}
	}
}

// Stop stops monitoring.
func (m *Monitor) Stop() {
	if m.stopMon != nil {
		m.stopMon()
	}
}

// Round is one round of monitoring.
func (m *Monitor) Round() error {
	logger := m.logger.Add("method", "Round")

	logger.Debug("monitoring round started")

	latestBlock, err := m.latestBlockNumber()
	if err != nil {
		return fmt.Errorf("could not get the latest block: %v", err)
	}

	queries, doneRound, err := m.NextRound(latestBlock)
	if err != nil {
		return fmt.Errorf("could not get filter logs queries: %v", err)
	}

	// No queries to proceed.
	if len(queries) == 0 {
		logger.Debug("no queries to proceed")
		return nil
	}

	var jobsToCreate []data.Job

	for _, query := range queries {
		logs, err := m.filterLogs(query)
		if err != nil {
			return fmt.Errorf("could not filter logs: %v", err)
		}
		for _, log := range logs {
			jEthLog := &data.JobEthLog{
				Block:  log.BlockNumber,
				Data:   log.Data,
				Topics: log.Topics,
				TxHash: data.HexFromBytes(log.TxHash.Bytes()),
			}

			logger.Add("topicHash", log.Topics[0].String(),
				"blockNumber", log.BlockNumber,
				"transactionHash", log.TxHash.String()).Debug(
				"received Ethereum log")
			jobs, err := m.JobsForLog(jEthLog)

			if err != nil {
				return fmt.Errorf("could not construct jobs for eth log %v: %v", jEthLog, err)
			}
			jobsToCreate = append(jobsToCreate, jobs...)
		}
	}

	return m.db.InTransaction(func(tx *reform.TX) error {
		for _, job := range jobsToCreate {
			job.CreatedBy = data.JobBCMonitor

			if err := m.Queue.Add(tx, &job); err != nil {
				return fmt.Errorf("could not insert a job: %v", err)
			}
		}

		return doneRound(tx)
	})
}

func (m *Monitor) latestBlockNumber() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.RequestTimeout)
	defer cancel()
	latestBlock, err := m.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, err
	}

	return latestBlock.Number.Uint64(), nil
}

func (m *Monitor) filterLogs(query ethereum.FilterQuery) ([]types.Log, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.RequestTimeout)
	defer cancel()
	logs, err := m.client.FilterLogs(ctx, query)
	if err != nil {
		return nil, err
	}
	return logs, nil
}
