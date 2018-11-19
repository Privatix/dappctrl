package monitor

import (
	"context"
	"encoding/json"

	"github.com/ethereum/go-ethereum"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
)

type queriesBuilderFunc func(from, to uint64) ([]ethereum.FilterQuery, error)

func (m *Monitor) queryLogsAndCreateJobs(
	builder queriesBuilderFunc, producers JobsProducers) error {
	logger := m.logger.Add("method", "queryLogsAndCreateJobs")

	ctx, cancel := context.WithTimeout(context.Background(),
		m.ethCallTimeout)
	defer cancel()
	from, to, err := m.rangeOfInterest(ctx)
	if err != nil || from >= to {
		return err
	}

	logger = logger.Add("from", from, "to", to)

	queries, err := builder(from, to)
	if err != nil {
		return err
	}

	jobsToCreate := make([]data.Job, 0)

	for _, query := range queries {
		logs, err := m.eth.FilterLogs(ctx, query)
		if err != nil {
			logger.Error(err.Error())
			return ErrFailedToFetchLogs
		}

		for _, log := range logs {
			jEthLog := &data.JobEthLog{
				Block:  log.BlockNumber,
				Data:   log.Data,
				Topics: log.Topics,
				TxHash: data.HexFromBytes(log.TxHash.Bytes()),
			}

			topicName := m.eventNameFromHash(log.Topics[0])

			logger.Add("topicName", topicName,
				"topicHash", log.Topics[0].String(),
				"blockNumber", log.BlockNumber,
				"transactionHash", log.TxHash.String()).Debug(
				"received Ethereum log")

			producerF, ok := producers[log.Topics[0]]
			if !ok {
				logger.Error("no producer function for event: " +
					log.Topics[0].Hex() + ". skipping")
				continue
			}
			jobs, err := producerF(jEthLog)
			if err != nil {
				return err
			}

			jobsToCreate = append(jobsToCreate, jobs...)
		}
	}

	return m.db.InTransaction(func(tx *reform.TX) error {
		for _, job := range jobsToCreate {
			job.CreatedBy = data.JobBCMonitor

			err = m.queue.Add(tx, &job)
			if err != nil {
				log := data.JobData{}
				json.Unmarshal(job.Data, &log)
				logger.Add("job", job,
					"jobEthLog", *log.EthLog).Error(err.Error())
				return err
			}
		}
		_, err := tx.Exec(`UPDATE settings SET value=$1 WHERE key=$2`,
			to, data.SettingLastProcessedBlock)
		return err
	})
}
