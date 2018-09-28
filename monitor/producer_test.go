package monitor_test

import (
	"bytes"
	"encoding/json"
	"math/big"
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

	agentHash := common.HexToHash(fxt.Channel.Agent)
	clientHash := common.HexToHash(fxt.Channel.Client)
	offeringHash := common.BytesToHash(data.TestToBytes(t, fxt.Offering.Hash))
	agentProducersMap := mon.AgentJobsProducers()
	clientProducersMap := mon.ClientJobsProducers()

	for _, tc := range []struct {
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
				Data: packEventData(t, "LogChannelToppedUp", fxt.Channel.Block, new(big.Int)),
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
				Data: packEventData(t, "LogChannelCloseRequested", fxt.Channel.Block, new(big.Int)),
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
				Data: packEventData(t, "LogUnCooperativeChannelClose", fxt.Channel.Block, new(big.Int)),
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
				Data: packEventData(t, "LogCooperativeChannelClose", fxt.Channel.Block, new(big.Int)),
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
		testProducedJobs(t, agentProducersMap, tc.l, tc.agentProduced)
		testProducedJobs(t, clientProducersMap, tc.l, tc.clientProduced)
	}
}

func testProducedJobs(t *testing.T, producers monitor.JobsProducers,
	l *data.JobEthLog, jobs []data.Job) {
	produced, err := producers[l.Topics[0]](l)
	util.TestExpectResult(t, "produceFunc", nil, err)
	if len(jobs) != len(produced) {
		t.Fatalf("wanted %v jobs, got: %v", len(jobs),
			len(produced))
	}
	for i, job := range produced {
		if (jobs[i].RelatedID != "" && jobs[i].RelatedID !=
			job.RelatedID) || jobs[i].RelatedType !=
			job.RelatedType || jobs[i].Type != job.Type {
			t.Fatal("wrong job produced ", job)
		}
		jData, _ := json.Marshal(&data.JobData{EthLog: l})
		if !bytes.Equal(jData, job.Data) {
			t.Fatal("log does not appear in job data")
		}
	}
}
