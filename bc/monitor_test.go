package bc

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

var (
	db     *reform.DB
	queue  Queue
	logger log.Logger
)

func TestMonitorAbstract(t *testing.T) {
	client := eth.NewTestEthBackend(common.HexToAddress("0x1"))

	client.Logs = []ethtypes.Log{
		{
			Topics: []common.Hash{common.HexToHash("0x23")},
		},
	}
	client.BlockNumber = new(big.Int).SetInt64(9999)
	testLog := &data.JobEthLog{
		Topics: []common.Hash{common.HexToHash("0x23")},
		TxHash: data.HexFromBytes(client.Logs[0].TxHash.Bytes()),
	}
	// Some random job just for test.
	testJobs := []data.Job{
		{
			RelatedType: data.JobOffering,
			RelatedID:   util.NewUUID(),
			Data:        []byte("{}"),
		},
	}

	mon := &Monitor{
		client: client,
		db:     db,
		logger: logger,
	}
	mon.Queue = queue

	mon.JobsForLog = func(logs *data.JobEthLog) ([]data.Job, error) {
		if !reflect.DeepEqual(testLog, logs) {
			t.Fatalf("unexpected log given to jobs maker, want: %v, got: %v", testLog, logs)
		}
		return testJobs, nil
	}

	mon.NextRound = func(latestBlock uint64) ([]ethereum.FilterQuery, func(*reform.TX) error, error) {
		// Return some query to force monitor to make query for ethereum logs.
		return []ethereum.FilterQuery{
			{},
		}, func(_ *reform.TX) error { return nil }, nil
	}

	// Force monitor round.
	if err := mon.Round(); err != nil {
		t.Fatal(err)
	}

	var j data.Job
	err := db.FindOneTo(&j, "related_id", testJobs[0].RelatedID)
	if err != nil {
		t.Fatal(err)
	}
	db.Delete(&j)
}
