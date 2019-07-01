package monitor_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/monitor"
	"github.com/privatix/dappctrl/util"
)

func packEventData(t *testing.T, ename string, args ...interface{}) []byte {
	ret, err := pscABI.Events[ename].Inputs.NonIndexed().Pack(args...)
	if err != nil {
		t.Fatal(err)
	}
	return ret
}

func TestJobsProducers(t *testing.T) {
	fxt := data.NewTestFixture(t, db)
	etx := &data.EthTx{
		ID:          util.NewUUID(),
		Hash:        data.HexFromBytes(randHash().Bytes()),
		RelatedID:   fxt.Channel.ID,
		RelatedType: data.JobChannel,
		Status:      data.TxMined,
		TxRaw:       []byte("{}"),
		Gas:         1,
		GasPrice:    1,
	}
	data.InsertToTestDB(t, db, etx)
	defer data.DeleteFromTestDB(t, db, etx)
	defer fxt.Close()

	agentHash := common.HexToHash(string(fxt.Channel.Agent))
	clientHash := common.HexToHash(string(fxt.Channel.Client))
	offeringHash := common.HexToHash(string(fxt.Offering.Hash))
	agentProducersMap := mon.AgentJobsProducers()
	clientProducersMap := mon.ClientJobsProducers()

	for i, tc := range []struct {
		l              *data.JobEthLog
		agentProduced  []data.Job
		clientProduced []data.Job
	}{
		{
			l: &data.JobEthLog{
				Topics: []common.Hash{
					eth.ServiceChannelCreated,
					agentHash,
					clientHash,
					offeringHash,
				},
				TxHash: etx.Hash,
			},
			agentProduced: []data.Job{
				{
					RelatedID:   "",
					RelatedType: data.JobChannel,
					Type:        data.JobAgentAfterChannelCreate,
				},
			},
			clientProduced: []data.Job{
				{
					RelatedID:   fxt.Channel.ID,
					RelatedType: data.JobChannel,
					Type:        data.JobClientAfterChannelCreate,
				},
				{
					RelatedID:   fxt.Offering.ID,
					RelatedType: data.JobOffering,
					Type:        data.JobDecrementCurrentSupply,
				},
			},
		},
		{
			l: &data.JobEthLog{
				Topics: []common.Hash{
					eth.ServiceChannelToppedUp,
					agentHash,
					clientHash,
					offeringHash,
				},
				Data: packEventData(t, "LogChannelToppedUp", fxt.Channel.Block, uint64(0)),
			},
			agentProduced: []data.Job{
				{
					RelatedID:   fxt.Channel.ID,
					RelatedType: data.JobChannel,
					Type:        data.JobAgentAfterChannelTopUp,
				},
			},
			clientProduced: []data.Job{
				{
					RelatedID:   fxt.Channel.ID,
					RelatedType: data.JobChannel,
					Type:        data.JobClientAfterChannelTopUp,
				},
			},
		},
		{
			l: &data.JobEthLog{
				Topics: []common.Hash{
					eth.ServiceChannelCloseRequested,
					agentHash,
					clientHash,
					offeringHash,
				},
				Data: packEventData(t, "LogChannelCloseRequested", fxt.Channel.Block, uint64(0)),
			},
			agentProduced: []data.Job{
				{
					RelatedID:   fxt.Channel.ID,
					RelatedType: data.JobChannel,
					Type:        data.JobAgentAfterUncooperativeCloseRequest,
				},
			},
			clientProduced: []data.Job{
				{
					RelatedID:   fxt.Channel.ID,
					RelatedType: data.JobChannel,
					Type:        data.JobClientAfterUncooperativeCloseRequest,
				},
			},
		},
		{
			l: &data.JobEthLog{
				Topics: []common.Hash{
					eth.ServiceOfferingCreated,
					agentHash,
					offeringHash,
				},
			},
			agentProduced: []data.Job{
				{
					RelatedID:   fxt.Offering.ID,
					RelatedType: data.JobOffering,
					Type:        data.JobAgentAfterOfferingMsgBCPublish,
				},
			},
			clientProduced: []data.Job{
				{
					RelatedID:   "",
					RelatedType: data.JobOffering,
					Type:        data.JobClientAfterOfferingMsgBCPublish,
				},
			},
		},
		{
			l: &data.JobEthLog{
				Topics: []common.Hash{
					eth.ServiceOfferingDeleted,
					agentHash,
					offeringHash,
				},
			},
			agentProduced: []data.Job{
				{
					RelatedID:   fxt.Offering.ID,
					RelatedType: data.JobOffering,
					Type:        data.JobAgentAfterOfferingDelete,
				},
			},
			clientProduced: []data.Job{
				{
					RelatedID:   fxt.Offering.ID,
					RelatedType: data.JobOffering,
					Type:        data.JobClientAfterOfferingDelete,
				},
			},
		},
		{
			l: &data.JobEthLog{
				Topics: []common.Hash{
					eth.ServiceOfferingPopedUp,
					agentHash,
					offeringHash,
				},
			},
			agentProduced: []data.Job{
				{
					RelatedID:   fxt.Offering.ID,
					RelatedType: data.JobOffering,
					Type:        data.JobAgentAfterOfferingPopUp,
				},
			},
			clientProduced: []data.Job{
				{
					RelatedID:   fxt.Offering.ID,
					RelatedType: data.JobOffering,
					Type:        data.JobClientAfterOfferingPopUp,
				},
			},
		},
		{
			l: &data.JobEthLog{
				Topics: []common.Hash{
					eth.ServiceUnCooperativeChannelClose,
					agentHash,
					clientHash,
					offeringHash,
				},
				Data: packEventData(t, "LogUnCooperativeChannelClose", fxt.Channel.Block, uint64(0)),
			},
			agentProduced: []data.Job{
				{
					RelatedID:   fxt.Channel.ID,
					RelatedType: data.JobChannel,
					Type:        data.JobAgentAfterUncooperativeClose,
				},
			},
			clientProduced: []data.Job{
				{
					RelatedID:   fxt.Channel.ID,
					RelatedType: data.JobChannel,
					Type:        data.JobClientAfterUncooperativeClose,
				},
				{
					RelatedType: data.JobChannel,
					Type:        data.JobClientRecordClosing,
				},
				{
					RelatedID:   fxt.Offering.ID,
					RelatedType: data.JobOffering,
					Type:        data.JobIncrementCurrentSupply,
				},
			},
		},
		{
			l: &data.JobEthLog{
				Topics: []common.Hash{
					eth.ServiceCooperativeChannelClose,
					agentHash,
					clientHash,
					offeringHash,
				},
				Data: packEventData(t, "LogCooperativeChannelClose", fxt.Channel.Block, uint64(0)),
			},
			agentProduced: []data.Job{
				{
					RelatedID:   fxt.Channel.ID,
					RelatedType: data.JobChannel,
					Type:        data.JobAgentAfterCooperativeClose,
				},
			},
			clientProduced: []data.Job{
				{
					RelatedID:   fxt.Channel.ID,
					RelatedType: data.JobChannel,
					Type:        data.JobClientAfterCooperativeClose,
				},
				{
					RelatedType: data.JobChannel,
					Type:        data.JobClientRecordClosing,
				},
				{
					RelatedID:   fxt.Offering.ID,
					RelatedType: data.JobOffering,
					Type:        data.JobIncrementCurrentSupply,
				},
			},
		},
		{
			l: &data.JobEthLog{
				Topics: []common.Hash{
					eth.TokenApproval,
					agentHash,
					randHash(),
				},
			},
			agentProduced: []data.Job{
				{
					RelatedID:   fxt.Account.ID,
					RelatedType: data.JobAccount,
					Type:        data.JobPreAccountAddBalance,
				},
			},
			clientProduced: []data.Job{
				{
					RelatedID:   fxt.Account.ID,
					RelatedType: data.JobAccount,
					Type:        data.JobPreAccountAddBalance,
				},
			},
		},
		{
			l: &data.JobEthLog{
				Topics: []common.Hash{
					eth.TokenTransfer,
					agentHash,
					pscAddr.Hash(),
				},
			},
			agentProduced: []data.Job{
				{
					RelatedID:   fxt.Account.ID,
					RelatedType: data.JobAccount,
					Type:        data.JobAfterAccountAddBalance,
				},
			},
			clientProduced: []data.Job{
				{
					RelatedID:   fxt.Account.ID,
					RelatedType: data.JobAccount,
					Type:        data.JobAfterAccountAddBalance,
				},
			},
		},
		{
			l: &data.JobEthLog{
				Topics: []common.Hash{
					eth.TokenTransfer,
					pscAddr.Hash(),
					agentHash,
				},
			},
			agentProduced: []data.Job{
				{
					RelatedID:   fxt.Account.ID,
					RelatedType: data.JobAccount,
					Type:        data.JobAfterAccountReturnBalance,
				},
			},
			clientProduced: []data.Job{
				{
					RelatedID:   fxt.Account.ID,
					RelatedType: data.JobAccount,
					Type:        data.JobAfterAccountReturnBalance,
				},
			},
		},
	} {
		t.Log("before", i)
		testProducedJobs(t, agentProducersMap, tc.l, nil, tc.agentProduced)
		t.Log("after", i)
		testProducedJobs(t, clientProducersMap, tc.l, nil, tc.clientProduced)
	}

	// Decrement/Increment current supply needs to find offering to decerement supply at.
	// If related Offering created job was not yet in db, it is most
	// likely in current list of jobs to create, which is in memory.
	// Test case below ensures decrement current supply correctly examines produced jobs
	// to find related offering id from job that is not yet in db.
	offeringHash = randHash()
	ethLog := &data.JobEthLog{
		Topics: []common.Hash{randHash(), randHash(), offeringHash},
	}
	jobData, _ := json.Marshal(&data.JobData{EthLog: ethLog})
	producingJobs := []data.Job{{
		Type:        data.JobClientAfterOfferingMsgBCPublish,
		RelatedID:   util.NewUUID(),
		RelatedType: data.JobOffering,
		Data:        jobData,
	}}
	for _, test := range []struct {
		log *data.JobEthLog
		ret []data.Job
	}{
		{
			log: &data.JobEthLog{
				Topics: []common.Hash{
					eth.ServiceChannelCreated,
					randHash(),
					randHash(),
					offeringHash,
				},
			},
			ret: []data.Job{
				{
					RelatedID:   producingJobs[0].RelatedID,
					RelatedType: data.JobOffering,
					Type:        data.JobDecrementCurrentSupply,
				},
			},
		},
		{
			log: &data.JobEthLog{
				Topics: []common.Hash{
					eth.ServiceCooperativeChannelClose,
					randHash(),
					randHash(),
					offeringHash,
				},
				Data: packEventData(t, "LogCooperativeChannelClose", fxt.Channel.Block, uint64(0)),
			},
			ret: []data.Job{
				{
					RelatedType: data.JobChannel,
					Type:        data.JobClientRecordClosing,
				},
				{
					RelatedID:   producingJobs[0].RelatedID,
					RelatedType: data.JobOffering,
					Type:        data.JobIncrementCurrentSupply,
				},
			},
		},
		{
			log: &data.JobEthLog{
				Topics: []common.Hash{
					eth.ServiceUnCooperativeChannelClose,
					randHash(),
					randHash(),
					offeringHash,
				},
				Data: packEventData(t, "LogCooperativeChannelClose", fxt.Channel.Block, uint64(0)),
			},
			ret: []data.Job{
				{
					RelatedType: data.JobChannel,
					Type:        data.JobClientRecordClosing,
				},
				{
					RelatedID:   producingJobs[0].RelatedID,
					RelatedType: data.JobOffering,
					Type:        data.JobIncrementCurrentSupply,
				},
			},
		},
	} {
		testProducedJobs(t, clientProducersMap, test.log, producingJobs, test.ret)
	}

	// It is possilbe to have both uncooperative and cooperative close logs related
	// to the same channel. This tests increment current supply job not created
	// for cooperative channel close if there were uncooperative channel close
	// and increment job.
	randHash1 := randHash()
	randHash2 := randHash()
	offeringHash = common.HexToHash(string(fxt.Offering.Hash))
	jobData, _ = json.Marshal(&data.JobData{EthLog: &data.JobEthLog{
		Topics: []common.Hash{
			eth.ServiceUnCooperativeChannelClose,
			randHash1,
			randHash2,
			offeringHash,
		},
		Data: packEventData(t, "LogUnCooperativeChannelClose", fxt.Channel.Block, uint64(0)),
	}})
	uncoopIncrementJob := &data.Job{
		ID:          util.NewUUID(),
		Type:        data.JobIncrementCurrentSupply,
		RelatedID:   fxt.Offering.ID,
		RelatedType: data.JobOffering,
		Data:        jobData,
		Status:      data.JobActive,
		CreatedBy:   data.JobBCMonitor,
	}
	data.InsertToTestDB(t, fxt.DB, uncoopIncrementJob)
	defer data.DeleteFromTestDB(t, fxt.DB, uncoopIncrementJob)
	testProducedJobs(t, clientProducersMap, &data.JobEthLog{
		Topics: []common.Hash{
			eth.ServiceCooperativeChannelClose,
			randHash1,
			randHash2,
			offeringHash,
		},
		Data: packEventData(t, "LogCooperativeChannelClose", fxt.Channel.Block, uint64(0)),
	}, nil, []data.Job{{
		RelatedType: data.JobChannel,
		Type:        data.JobClientRecordClosing,
	}})
}

func testProducedJobs(t *testing.T, producers monitor.JobsProducers,
	l *data.JobEthLog, alreadyProducedJobs, jobs []data.Job) {
	t.Helper()

	produced, err := producers[l.Topics[0]](l, alreadyProducedJobs)
	util.TestExpectResult(t, "produceFunc", nil, err)
	if len(jobs) != len(produced) {
		t.Fatalf("wanted %v jobs, got: %v", len(jobs),
			len(produced))
	}
	for i, job := range produced {
		if jobs[i].Type == data.JobClientRecordClosing {
			if jobs[i].RelatedType != job.RelatedType ||
				jobs[i].Type != job.Type {
				t.Fatal("wrong job produced ", job)
			}
			continue
		}

		if (jobs[i].RelatedID != "" && jobs[i].RelatedID !=
			job.RelatedID) || jobs[i].RelatedType !=
			job.RelatedType || jobs[i].Type != job.Type {
			t.Fatal("wrong job produced ", jobs[i].Type)
		}
		jData, _ := json.Marshal(&data.JobData{EthLog: l})
		if !bytes.Equal(jData, job.Data) {
			t.Fatal("log does not appear in job data")
		}
	}
}
