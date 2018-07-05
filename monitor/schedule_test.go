// +build !nomonitortest

package monitor

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/util"
)

var (
	LogChannelCreated            = "LogChannelCreated"
	LogChannelTopUp              = "LogChannelToppedUp"
	LogChannelCloseRequested     = "LogChannelCloseRequested"
	LogOfferingCreated           = "LogOfferingCreated"
	LogOfferingDeleted           = "LogOfferingDeleted"
	LogOfferingPopedUp           = "LogOfferingPopedUp"
	LogCooperativeChannelClose   = "LogCooperativeChannelClose"
	LogUnCooperativeChannelClose = "LogUnCooperativeChannelClose"
)

func TestMonitorSchedule(t *testing.T) {
	defer cleanDB(t)

	setMaxRetryKey(t)

	mon, queue, _ := newTestObjects(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := newMockTicker()
	mon.start(ctx, conf.BlockMonitor.Timeout, nil, ticker.C)

	td := generateTestData(t)

	scheduleTest(t, td, queue, ticker, mon)
}

func toHashes(t *testing.T, topics []interface{}) []common.Hash {
	hashes := make([]common.Hash, len(topics))
	if len(topics) > 0 {
		hashes[0] = common.HexToHash(topics[0].(string))
	}

	for i, topic := range topics[1:] {
		switch v := topic.(type) {
		case string:
			if bs, err := data.ToBytes(v); err == nil {
				hashes[i+1] = common.BytesToHash(bs)
			} else {
				t.Fatal(err)
			}
		case int:
			hashes[i+1] = common.BigToHash(big.NewInt(int64(v)))
		case common.Address:
			hashes[i+1] = v.Hash()
		case common.Hash:
			hashes[i+1] = v
		default:
			t.Fatalf("unsupported type %T as topic", topic)
		}
	}

	return hashes
}

func pscEvent(t *testing.T, mon *Monitor, customTxHash, logName string,
	topics, nonIndexedArgs []interface{}) *data.EthLog {
	ev, ok := mon.pscABI.Events[logName]
	if !ok {
		t.Fatal("events with this name do not exist")
	}

	top := append([]interface{}{}, ev.Id().String())
	top = append(top, topics...)

	el := insertEvent(t, db, nextBlock(), 0, top...)

	if customTxHash != "" {
		el.TxHash = customTxHash
		data.SaveToTestDB(t, db, el)
	}

	if len(nonIndexedArgs) != 0 {
		bs, err := ev.Inputs.NonIndexed().Pack(nonIndexedArgs...)
		if err != nil {
			t.Fatal(err)
		}
		el.Data = data.FromBytes(bs)

		if err := db.Save(el); err != nil {
			t.Fatal(err)
		}
	}

	return el
}

func insertEvent(t *testing.T, db *reform.DB, blockNumber uint64,
	failures uint64, topics ...interface{}) *data.EthLog {
	el := &data.EthLog{
		ID:          util.NewUUID(),
		TxHash:      data.FromBytes(genRandData(32)),
		TxStatus:    txMinedStatus,
		BlockNumber: blockNumber,
		Addr:        data.FromBytes(pscAddr.Bytes()),
		Data:        data.FromBytes(genRandData(32)),
		Topics:      toHashes(t, topics),
		Failures:    failures,
	}
	if err := db.Insert(el); err != nil {
		t.Fatalf("failed to insert a log event into db: %v", err)
	}

	return el
}

func setIsAgent(t *testing.T, agent bool) {
	var val string
	if agent {
		val = "true"
	} else {
		val = "false"
	}

	setting := &data.Setting{Key: data.IsAgentKey, Value: val,
		Name: "user role is agent"}
	data.SaveToTestDB(t, db, setting)
}

func agentSchedule(t *testing.T, td *testData, queue *mockQueue,
	ticker *mockTicker, mon *Monitor) {
	setIsAgent(t, true)

	// LogChannelCreated good
	pscEvent(t, mon, txHash, LogChannelCreated, []interface{}{
		td.addr[0],          // agent
		someAddress,         // client
		td.offering[0].Hash, // offering
	}, nil)

	queue.expect(data.JobAgentAfterChannelCreate, func(j *data.Job) bool {
		return j.Type == data.JobAgentAfterChannelCreate
	})

	// LogChannelCreated ignore
	pscEvent(t, mon, "", LogChannelCreated, []interface{}{
		someAddress,         // agent
		td.addr[0],          // client
		td.offering[0].Hash, // offering
	}, nil)

	// LogChannelToppedUp good
	pscEvent(t, mon, "", LogChannelTopUp, []interface{}{
		td.addr[0],          // agent
		someAddress,         // client
		td.offering[0].Hash, // offering
	}, []interface{}{uint32(td.channel[0].Block), new(big.Int)})

	queue.expect(data.JobAgentAfterChannelTopUp, func(j *data.Job) bool {
		return j.Type == data.JobAgentAfterChannelTopUp
	})

	// LogChannelToppedUp ignored
	// channel does not exist, thus event ignored
	pscEvent(t, mon, "", LogChannelTopUp, []interface{}{
		td.addr[0],          // agent
		td.addr[1],          // client
		td.offering[1].Hash, // offering
	}, []interface{}{uint32(td.channel[1].Block), new(big.Int)})

	// LogChannelCloseRequested good
	pscEvent(t, mon, "", LogChannelCloseRequested,
		[]interface{}{
			td.addr[0],          // agent
			someAddress,         // client
			td.offering[0].Hash, // offering
		}, []interface{}{uint32(td.channel[0].Block), new(big.Int)})

	queue.expect(data.JobAgentAfterUncooperativeCloseRequest,
		func(j *data.Job) bool {
			return j.Type ==
				data.JobAgentAfterUncooperativeCloseRequest
		})

	// LogChannelCloseRequested ignored
	pscEvent(t, mon, "", LogChannelCloseRequested,
		[]interface{}{
			td.addr[0],          // agent
			someAddress,         // client
			td.offering[0].Hash, // offering
		}, []interface{}{uint32(td.channel[1].Block), new(big.Int)})

	// LogOfferingCreated good
	pscEvent(t, mon, "", LogOfferingCreated, []interface{}{
		td.addr[0],          // agent
		td.offering[0].Hash, // offering hash
		minDepositVal,       // min deposit
	}, nil)

	queue.expect(data.JobAgentAfterOfferingMsgBCPublish,
		func(j *data.Job) bool {
			return j.Type ==
				data.JobAgentAfterOfferingMsgBCPublish
		})

	// LogOfferingCreated good
	pscEvent(t, mon, "", LogOfferingCreated, []interface{}{
		td.addr[1],          // agent
		td.offering[1].Hash, // offering
		minDepositVal,       // min deposit
	}, nil)

	queue.expect(data.JobAgentAfterOfferingMsgBCPublish,
		func(j *data.Job) bool {
			return j.Type ==
				data.JobAgentAfterOfferingMsgBCPublish
		})

	// LogOfferingCreated good
	pscEvent(t, mon, "", LogOfferingCreated, []interface{}{
		td.addr[1],          // agent
		td.offering[2].Hash, // offering
		minDepositVal,       // min deposit
	}, nil)

	queue.expect(data.JobAgentAfterOfferingMsgBCPublish,
		func(j *data.Job) bool {
			return j.Type ==
				data.JobAgentAfterOfferingMsgBCPublish
		})

	// LogOfferingDeleted good
	// TODO(maxim) implementation after afterOfferingDelete job implementation

	// LogOfferingPopedUp good
	pscEvent(t, mon, "", LogOfferingPopedUp, []interface{}{
		someAddress,         // agent
		td.offering[0].Hash, // offering hash
	}, nil)

	queue.expect(data.JobAgentAfterOfferingMsgBCPublish,
		func(j *data.Job) bool {
			return j.Type ==
				data.JobAgentAfterOfferingMsgBCPublish
		})

	// LogCooperativeChannelClose good
	pscEvent(t, mon, "", LogCooperativeChannelClose,
		[]interface{}{
			td.addr[0],          // agent
			someAddress,         // client
			td.offering[0].Hash, // offering
		}, []interface{}{uint32(td.channel[0].Block), new(big.Int)})

	queue.expect(data.JobAgentAfterCooperativeClose, func(j *data.Job) bool {
		return j.Type == data.JobAgentAfterCooperativeClose
	})

	// LogCooperativeChannelClose ignored
	pscEvent(t, mon, "", LogCooperativeChannelClose,
		[]interface{}{
			td.addr[0],          // agent
			someAddress,         // client
			td.offering[0].Hash, // offering
		}, []interface{}{uint32(td.channel[1].Block), new(big.Int)})

	// LogUnCooperativeChannelClose good
	pscEvent(t, mon, "", LogUnCooperativeChannelClose,
		[]interface{}{
			td.addr[0],          // agent
			someAddress,         // client
			td.offering[0].Hash, // offering
		}, []interface{}{uint32(td.channel[0].Block), new(big.Int)})

	queue.expect(data.JobAgentAfterUncooperativeClose,
		func(j *data.Job) bool {
			return j.Type ==
				data.JobAgentAfterUncooperativeClose
		})

	// LogUnCooperativeChannelClose ignored
	pscEvent(t, mon, "", LogUnCooperativeChannelClose,
		[]interface{}{
			someAddress,         // agent
			someAddress,         // client
			td.offering[0].Hash, // offering
		}, []interface{}{uint32(td.channel[1].Block), new(big.Int)})

	// Tick here on purpose, so that not all events are ignored because
	// the offering's been deleted.
	ticker.tick()
	queue.awaitCompletion(time.Second)
}

func clientSchedule(t *testing.T, td *testData, queue *mockQueue,
	ticker *mockTicker, mon *Monitor) {
	setIsAgent(t, false)

	// LogChannelCreated good
	pscEvent(t, mon, txHash, LogChannelCreated, []interface{}{
		someAddress,         // agent
		td.addr[1],          // client
		td.offering[1].Hash, // offering
	}, nil)

	queue.expect(data.JobClientAfterChannelCreate, func(j *data.Job) bool {
		return j.Type == data.JobClientAfterChannelCreate
	})

	// LogChannelCreated ignore
	pscEvent(t, mon, "", LogChannelCreated, []interface{}{
		someAddress,         // agent
		td.addr[0],          // client
		td.offering[1].Hash, // offering
	}, nil)

	// LogChannelToppedUp good
	pscEvent(t, mon, "", LogChannelTopUp, []interface{}{
		someAddress,         // agent
		td.addr[1],          // client
		td.offering[1].Hash, // offering
	}, []interface{}{uint32(td.channel[1].Block), new(big.Int)})

	queue.expect(data.JobClientAfterChannelTopUp, func(j *data.Job) bool {
		return j.Type == data.JobClientAfterChannelTopUp
	})

	// LogChannelToppedUp ignore
	pscEvent(t, mon, "", LogChannelTopUp, []interface{}{
		td.addr[1],          // agent
		td.addr[0],          // client
		td.offering[1].Hash, // offering
	}, []interface{}{uint32(td.channel[1].Block), new(big.Int)})

	// LogChannelCloseRequested good
	pscEvent(t, mon, "", LogChannelCloseRequested,
		[]interface{}{
			someAddress,         // agent
			td.addr[1],          // client
			td.offering[1].Hash, // offering
		}, []interface{}{uint32(td.channel[1].Block), new(big.Int)})

	queue.expect(data.JobClientAfterUncooperativeCloseRequest,
		func(j *data.Job) bool {
			return j.Type ==
				data.JobClientAfterUncooperativeCloseRequest
		})

	// LogChannelCloseRequested ignored
	pscEvent(t, mon, "", LogChannelCloseRequested,
		[]interface{}{
			someAddress,         // agent
			td.addr[0],          // client
			td.offering[1].Hash, // offering
		}, []interface{}{uint32(td.channel[1].Block), new(big.Int)})

	// LogOfferingCreated good
	pscEvent(t, mon, "", LogOfferingCreated, []interface{}{
		td.addr[0],          // agent
		td.offering[1].Hash, // offering
		minDepositVal,       // min deposit
	}, nil)

	queue.expect(data.JobClientAfterOfferingMsgBCPublish,
		func(j *data.Job) bool {
			return j.Type ==
				data.JobClientAfterOfferingMsgBCPublish
		})

	// LogOfferingCreated good
	pscEvent(t, mon, "", LogOfferingCreated, []interface{}{
		td.addr[1],          // agent
		td.offering[1].Hash, // offering
		minDepositVal,       // min deposit
	}, nil)

	queue.expect(data.JobClientAfterOfferingMsgBCPublish,
		func(j *data.Job) bool {
			return j.Type ==
				data.JobClientAfterOfferingMsgBCPublish
		})

	// LogOfferingCreated good
	pscEvent(t, mon, "", LogOfferingCreated, []interface{}{
		td.addr[1],          // agent
		td.offering[2].Hash, // offering
		minDepositVal,       // min deposit
	}, nil)

	queue.expect(data.JobClientAfterOfferingMsgBCPublish,
		func(j *data.Job) bool {
			return j.Type ==
				data.JobClientAfterOfferingMsgBCPublish
		})

	// LogOfferingDeleted good
	// TODO(maxim) implementation after afterOfferingDelete job implementation

	// LogOfferingPopedUp good
	pscEvent(t, mon, "", LogOfferingPopedUp, []interface{}{
		someAddress,         // agent
		td.offering[2].Hash, // offering hash
	}, nil)

	queue.expect(data.JobClientAfterOfferingPopUp,
		func(j *data.Job) bool {
			return j.Type ==
				data.JobClientAfterOfferingPopUp
		})

	// LogCooperativeChannelClose good
	pscEvent(t, mon, "", LogCooperativeChannelClose,
		[]interface{}{
			someAddress,         // agent
			td.addr[1],          // client
			td.offering[1].Hash, // offering
		}, []interface{}{uint32(td.channel[1].Block), new(big.Int)})

	queue.expect(data.JobClientAfterCooperativeClose, func(j *data.Job) bool {
		return j.Type == data.JobClientAfterCooperativeClose
	})

	// LogCooperativeChannelClose ignored
	pscEvent(t, mon, "", LogCooperativeChannelClose,
		[]interface{}{
			someAddress,         // agent
			td.addr[0],          // client
			td.offering[1].Hash, // offering
		}, []interface{}{uint32(td.channel[0].Block), new(big.Int)})

	// LogUnCooperativeChannelClose good
	pscEvent(t, mon, "", LogUnCooperativeChannelClose,
		[]interface{}{
			someAddress,         // agent
			td.addr[1],          // client
			td.offering[1].Hash, // offering
		}, []interface{}{uint32(td.channel[1].Block), new(big.Int)})

	queue.expect(data.JobClientAfterUncooperativeClose,
		func(j *data.Job) bool {
			return j.Type ==
				data.JobClientAfterUncooperativeClose
		})

	// LogUnCooperativeChannelClose ignored
	pscEvent(t, mon, "", LogUnCooperativeChannelClose,
		[]interface{}{
			someAddress,         // agent
			td.addr[1],          // client
			td.offering[1].Hash, // offering
		}, []interface{}{uint32(td.channel[0].Block), new(big.Int)})

	// Tick here on purpose, so that not all events are ignored because
	// the offering's been deleted.
	ticker.tick()
	queue.awaitCompletion(time.Second)
}

func commonSchedule(t *testing.T, td *testData, queue *mockQueue,
	ticker *mockTicker, agent bool) {
	setIsAgent(t, agent)

	insertEvent(t, db, nextBlock(), 0, eth.EthTokenApproval,
		td.addr[0], pscAddr, 123)

	queue.expect(data.JobPreAccountAddBalance, func(j *data.Job) bool {
		return j.Type == data.JobPreAccountAddBalance
	})

	insertEvent(t, db, nextBlock(), 0, eth.EthTokenTransfer,
		td.addr[0], someAddress, 123)

	queue.expect(data.JobAfterAccountAddBalance, func(j *data.Job) bool {
		return j.Type == data.JobAfterAccountAddBalance
	})

	insertEvent(t, db, nextBlock(), 0, eth.EthTokenTransfer,
		someAddress, td.addr[0], 123)

	queue.expect(data.JobAfterAccountAddBalance, func(j *data.Job) bool {
		return j.Type == data.JobAfterAccountAddBalance
	})

	// Tick here on purpose, so that not all events are ignored because
	// the offering's been deleted.
	ticker.tick()
	queue.awaitCompletion(time.Second)
}

func scheduleTest(t *testing.T, td *testData, queue *mockQueue,
	ticker *mockTicker, mon *Monitor) {
	commonSchedule(t, td, queue, ticker, true)
	commonSchedule(t, td, queue, ticker, false)
	agentSchedule(t, td, queue, ticker, mon)
	clientSchedule(t, td, queue, ticker, mon)
}
