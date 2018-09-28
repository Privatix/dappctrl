package monitor_test

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

var byLogHashSQLTail = "WHERE data->'ethereumLog'->>'transactionHash'=$1"

func randHash() common.Hash {
	randbytes := make([]byte, common.HashLength)
	rand.Read(randbytes)
	return common.BytesToHash(randbytes[:])
}

func randJobForLog(l *data.JobEthLog) data.Job {
	jData, _ := json.Marshal(&data.JobData{EthLog: l})
	return data.Job{
		ID:          util.NewUUID(),
		Type:        "some type",
		RelatedType: data.JobChannel,
		RelatedID:   util.NewUUID(),
		Data:        jData,
	}
}

func randEthLog() ethtypes.Log {
	return ethtypes.Log{
		Topics: []common.Hash{randHash()},
		TxHash: randHash(),
	}
}

func buildOneQuery(_, _ uint64) ([]ethereum.FilterQuery, error) {
	return []ethereum.FilterQuery{{}}, nil
}
func TestQueryLogsAndCreateJobs(t *testing.T) {
	cleanup := blockSettings(t, 10, 10, 0, 0)
	defer cleanup()

	logs := []ethtypes.Log{randEthLog(), randEthLog()}
	ethClient.FilterLogsResult = logs
	ethClient.HeaderByNumberResult = 100

	producers := map[common.Hash]func(l *data.JobEthLog) ([]data.Job, error){
		logs[0].Topics[0]: func(l *data.JobEthLog) ([]data.Job, error) {
			return []data.Job{
				randJobForLog(l),
				randJobForLog(l),
			}, nil
		},
		logs[1].Topics[0]: func(l *data.JobEthLog) ([]data.Job, error) {
			return []data.Job{
				randJobForLog(l),
			}, nil
		},
	}

	err := mon.QueryLogsAndCreateJobs(buildOneQuery, producers)
	util.TestExpectResult(t, "QueryLogsAndCreateJobs", nil, err)

	// Test all produced jobs created.
	jobs, err := db.SelectAllFrom(data.JobTable,
		byLogHashSQLTail, data.HexFromBytes(logs[0].TxHash.Bytes()))
	util.TestExpectResult(t, "SelectOneFrom", nil, err)

	for _, job := range jobs {
		data.DeleteFromTestDB(t, db, job.(*data.Job))
	}
	if len(jobs) != 2 {
		t.Fatal("produced 2 jobs, created only: ", len(jobs))
	}

	// Test all log events proccessed.
	job := data.Job{}
	err = db.SelectOneTo(&job, byLogHashSQLTail,
		data.HexFromBytes(logs[1].TxHash.Bytes()))
	util.TestExpectResult(t, "SelectOneFrom", nil, err)
	defer data.DeleteFromTestDB(t, db, &job)

	// Test job created_by set correctly.
	if data.JobBCMonitor != job.CreatedBy {
		t.Fatalf("wanted: %v, got: %v", data.JobBCMonitor, job.CreatedBy)
	}

	lastBlock, _ := data.ReadUintSetting(db.Querier,
		data.SettingLastProcessedBlock)
	// Given that min confirmations is 0, last proccessed must be equal to
	// the last block number.
	if uint64(lastBlock) != ethClient.HeaderByNumberResult {
		t.Fatalf("wrong last proccessed block, wanted: %v, got: %v",
			ethClient.HeaderByNumberResult, lastBlock)
	}
}

func TestNoJobCreatedIfAnyProducerFails(t *testing.T) {
	cleanup := blockSettings(t, 10, 10, 0, 0)
	defer cleanup()

	logs := []ethtypes.Log{randEthLog(), randEthLog()}
	ethClient.FilterLogsResult = logs
	ethClient.HeaderByNumberResult = 100

	testErr := fmt.Errorf("test err")
	producers := map[common.Hash]func(l *data.JobEthLog) ([]data.Job, error){
		logs[0].Topics[0]: func(l *data.JobEthLog) ([]data.Job, error) {
			return []data.Job{
				randJobForLog(l),
			}, nil
		},
		logs[1].Topics[0]: func(l *data.JobEthLog) ([]data.Job, error) {
			return nil, testErr
		},
	}

	err := mon.QueryLogsAndCreateJobs(buildOneQuery, producers)
	util.TestExpectResult(t, "QueryLogsAndCreateJobs", testErr, err)

	for _, log := range logs {
		job := data.Job{}
		err = db.SelectOneTo(&job, byLogHashSQLTail, log.TxHash.Hex())
		util.TestExpectResult(t, "SelectOneFrom", sql.ErrNoRows, err)
	}
}

func TestNothingToQuery(t *testing.T) {
	cleanup := blockSettings(t, 10, 10, 10, 0)
	defer cleanup()

	log := randEthLog()
	ethClient.FilterLogsResult = []ethtypes.Log{log}
	ethClient.HeaderByNumberResult = 10

	producers := map[common.Hash]func(l *data.JobEthLog) ([]data.Job, error){
		log.Topics[0]: func(*data.JobEthLog) ([]data.Job, error) {
			return []data.Job{{}}, nil
		},
	}

	err := mon.QueryLogsAndCreateJobs(buildOneQuery, producers)
	util.TestExpectResult(t, "QueryLogsAndCreateJobs", nil, err)
}
