package worker

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/messages"
	"github.com/privatix/dappctrl/messages/ept"
	"github.com/privatix/dappctrl/util"
)

func TestClientPreChannelCreate(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientPreChannelCreate, data.JobChannel)
	defer fxt.Close()

	fxt.job.RelatedType = data.JobChannel
	fxt.job.RelatedID = util.NewUUID()
	fxt.setJobData(t, ClientPreChannelCreateData{
		Account:  fxt.Account.ID,
		Oferring: fxt.Offering.ID,
	})

	minDeposit := fxt.Offering.UnitPrice*fxt.Offering.MinUnits +
		fxt.Offering.SetupPrice
	env.ethBack.balancePSC = big.NewInt(int64(minDeposit - 1))
	util.TestExpectResult(t, "Job run", ErrNotEnoughBalance,
		env.worker.ClientPreChannelCreate(fxt.job))

	env.ethBack.balancePSC = big.NewInt(int64(minDeposit))
	util.TestExpectResult(t, "Job run", ErrNoSupply,
		env.worker.ClientPreChannelCreate(fxt.job))

	issued := time.Now()
	env.ethBack.offerSupply = 1
	runJob(t, env.worker.ClientPreChannelCreate, fxt.job)

	var tx data.EthTx
	ch := data.Channel{ID: fxt.job.RelatedID}
	data.FindInTestDB(t, db, &tx, "related_id", ch.ID)
	data.ReloadFromTestDB(t, db, &ch)
	defer data.DeleteFromTestDB(t, db, &ch, &tx)

	if ch.Agent != fxt.Offering.Agent ||
		ch.Client != fxt.Account.EthAddr ||
		ch.Offering != fxt.Offering.ID || ch.Block != 0 ||
		ch.ChannelStatus != data.ChannelPending ||
		ch.ServiceStatus != data.ServicePending ||
		ch.TotalDeposit != minDeposit {
		t.Fatalf("wrong channel content")
	}

	if tx.Method != "CreateChannel" || tx.Status != data.TxSent ||
		tx.JobID == nil || *tx.JobID != fxt.job.ID ||
		tx.Issued.Before(issued) || tx.Issued.After(time.Now()) ||
		tx.AddrFrom != fxt.Account.EthAddr ||
		tx.AddrTo != fxt.Offering.Agent ||
		tx.Nonce == nil || *tx.Nonce != fmt.Sprint(testTXNonce) ||
		tx.GasPrice != uint64(testTXGasPrice) ||
		tx.Gas != uint64(testTXGasLimit) ||
		tx.RelatedType != data.JobChannel {
		t.Fatalf("wrong transaction content")
	}
}

func TestClientAfterChannelCreate(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientAfterChannelCreate, data.JobChannel)
	defer fxt.Close()

	fxt.Channel.ServiceStatus = data.ServicePending
	data.SaveToTestDB(t, db, fxt.Channel)

	runJob(t, env.worker.ClientAfterChannelCreate, fxt.job)
	env.fakeSOMC.WriteWaitForEndpoint(t, fxt.Channel.ID, nil)
	time.Sleep(conf.JobHandlerTest.ReactionDelay * time.Millisecond)

	data.ReloadFromTestDB(t, db, fxt.Channel)
	if fxt.Channel.ChannelStatus != data.ChannelActive {
		t.Fatalf("expected %s service status, but got %s",
			data.ChannelActive, fxt.Channel.ChannelStatus)
	}

	var job data.Job
	err := env.worker.db.SelectOneTo(&job,
		"WHERE related_id = $1 AND id != $2",
		fxt.Channel.ID, fxt.job.ID)
	if err != nil {
		t.Fatalf("no new job created")
	}
	defer data.DeleteFromTestDB(t, db, &job)
}

func swapAgentWithClient(t *testing.T, fxt *workerTestFixture) {
	addr := fxt.Channel.Client
	fxt.Channel.Client = fxt.Channel.Agent
	fxt.Channel.Agent = addr

	fxt.User.PublicKey = fxt.Account.PublicKey

	data.SaveToTestDB(t, db, fxt.Channel, fxt.User)
}

func sealMessage(t *testing.T, env *workerTest,
	fxt *workerTestFixture, msg *ept.Message) []byte {
	mdata, _ := json.Marshal(&msg)

	pub, err := data.ToBytes(fxt.User.PublicKey)
	util.TestExpectResult(t, "Decode pub", nil, err)

	key, err := env.worker.key(fxt.Account.PrivateKey)
	util.TestExpectResult(t, "Get key", nil, err)

	sealed, err := messages.AgentSeal(mdata, pub, key)
	util.TestExpectResult(t, "AgentSeal", nil, err)

	return sealed
}

func TestClientPreEndpointMsgSOMCGet(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientPreEndpointMsgSOMCGet, data.JobChannel)
	defer fxt.Close()

	swapAgentWithClient(t, fxt)

	msg := ept.Message{
		TemplateHash:           "test-hash",
		Username:               util.NewUUID(),
		Password:               "test-password",
		PaymentReceiverAddress: "1.2.3.4:5678",
		ServiceEndpointAddress: "test-endpoint-addr",
		AdditionalParams: map[string]string{
			"test-key": "test-value",
		},
	}
	sealed := sealMessage(t, env, fxt, &msg)

	fxt.setJobData(t, sealed)
	runJob(t, env.worker.ClientPreEndpointMsgSOMCGet, fxt.job)

	var endp data.Endpoint
	data.SelectOneFromTestDBTo(t, db, &endp,
		"WHERE channel = $1 AND id != $2",
		fxt.Channel.ID, fxt.Endpoint.ID)
	defer data.DeleteFromTestDB(t, db, &endp)

	params, _ := json.Marshal(msg.AdditionalParams)
	if endp.Template != fxt.Offering.Template ||
		strings.Trim(endp.Hash, " ") != msg.TemplateHash ||
		endp.RawMsg != data.FromBytes(sealed) ||
		endp.Status != data.MsgUnpublished ||
		endp.PaymentReceiverAddress == nil ||
		*endp.PaymentReceiverAddress != msg.PaymentReceiverAddress ||
		endp.ServiceEndpointAddress == nil ||
		*endp.ServiceEndpointAddress != msg.ServiceEndpointAddress ||
		endp.Username == nil || *endp.Username != msg.Username ||
		endp.Password == nil || *endp.Password != msg.Password ||
		string(endp.AdditionalParams) != string(params) {
		t.Fatalf("bad endpoint content")
	}
}

func TestClientAfterEndpointMsgSOMCGet(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientAfterEndpointMsgSOMCGet, data.JobEndpoint)
	defer fxt.Close()

	err := fmt.Errorf("some error")
	env.worker.deployConfig = func(db *reform.DB, endpoint, dir string) error {
		return err
	}

	util.TestExpectResult(t, "run job", err,
		env.worker.ClientAfterEndpointMsgSOMCGet(fxt.job))

	err = nil
	runJob(t, env.worker.ClientAfterEndpointMsgSOMCGet, fxt.job)

	var endp data.Endpoint
	data.FindInTestDB(t, db, &endp, "id", fxt.Endpoint.ID)

	var ch data.Channel
	data.FindInTestDB(t, db, &ch, "id", fxt.Channel.ID)

	if endp.Status != data.MsgChPublished ||
		ch.ServiceStatus != data.ServiceSuspended {
		t.Fatalf("bad endpoint or channel status")
	}
}

func TestClientPreChannelTopUp(t *testing.T) {
	t.Skip("TODO")
	// 1. Check sufficient internal balance exists PSC.BalanceOf()
	// 2. PSC.topUpChannel()
}

func TestClientAfterChannelTopUp(t *testing.T) {
	t.Skip("TODO")
	// 1. Add deposit to channels.total_deposit
}

func TestClientPreUncooperativeCloseRequest(t *testing.T) {
	t.Skip("TODO")
	// 1. PSC.uncooperativeClose
	// 2. set ch_status="wait_challenge"
}

func TestClientAfterUncooperativeCloseRequest(t *testing.T) {
	t.Skip("TODO")
	// 1. set ch_status="in_challenge"
	// 2. "preUncooperativeClose" with delay
	// 3. "preServiceTerminate"
}

func TestClientPreUncooperativeClose(t *testing.T) {
	t.Skip("TODO")
	// 1. Check challenge_period ended.
	// 2. PSC.settle()
	// 3. set ch_status="wait_uncoop"
}

func TestClientAfterUncooperativeClose(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientAfterUncooperativeClose, data.JobChannel)
	defer fxt.Close()

	var job data.Job
	if err := env.worker.db.SelectOneTo(&job, "WHERE related_id = $1",
		fxt.Channel.ID); err != nil {
		t.Fatal(err)
	}
	defer data.DeleteFromTestDB(t, db, &job)

	fxt.Channel.ChannelStatus = data.ChannelWaitUncoop
	data.SaveToTestDB(t, db, fxt.Channel)

	runJob(t, env.worker.ClientAfterUncooperativeClose, fxt.job)

	var ch data.Channel
	if err := data.FindByPrimaryKeyTo(db.Querier, &ch,
		fxt.Channel.ID); err != nil {
		t.Fatal(err)
	}

	if ch.ChannelStatus != data.ChannelClosedUncoop {
		t.Fatalf("expected %s channel status, but got %s",
			data.ChannelClosedUncoop, fxt.Channel.ChannelStatus)
	}
}

func TestClientAfterCooperativeClose(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientAfterCooperativeClose, data.JobChannel)
	defer fxt.Close()

	var job data.Job
	if err := env.worker.db.SelectOneTo(&job, "WHERE related_id = $1",
		fxt.Channel.ID); err != nil {
		t.Fatal(err)
	}
	defer data.DeleteFromTestDB(t, db, &job)

	fxt.Channel.ChannelStatus = data.ChannelWaitUncoop
	data.SaveToTestDB(t, db, fxt.Channel)

	runJob(t, env.worker.ClientAfterCooperativeClose, fxt.job)

	var jobTerm data.Job
	if err := env.worker.db.SelectOneTo(&jobTerm,
		"WHERE related_id = $1 AND id != $2",
		fxt.Channel.ID, fxt.job.ID); err != nil {
		t.Fatal(err)
	}
	defer data.DeleteFromTestDB(t, db, &jobTerm)

	if jobTerm.Type != data.JobClientPreServiceTerminate {
		t.Fatal("bad job type")
	}

	var ch data.Channel
	if err := data.FindByPrimaryKeyTo(db.Querier, &ch,
		fxt.Channel.ID); err != nil {
		t.Fatal(err)
	}

	if ch.ChannelStatus != data.ChannelClosedCoop {
		t.Fatalf("expected %s service status, but got %s",
			data.ChannelClosedCoop, fxt.Channel.ChannelStatus)
	}
}

func TestClientPreServiceTerminate(t *testing.T) {
	t.Skip("TODO")
	// 1. svc_status="Terminated"
}

func TestClientAfterOfferingMsgBCPublish(t *testing.T) {
	t.Skip("TODO")
	// 1. "preOfferingMsgSOMCGet"
}

func TestClientOfferingMsgSOMCGet(t *testing.T) {
	t.Skip("TODO")
	// 1. Get OfferingMessage from SOMC
	// 2. set msg_status="msg_channel_published"
}

func TestClientPreAccountAddBalanceApprove(t *testing.T) {
	t.Skip("TODO")
	// 1. PTC.balanceOf()
	// 2. PTC.approve()
}

func TestClientPreAccountAddBalance(t *testing.T) {
	t.Skip("TODO")
	// 1. PSC.addBalanceERC20()
}

func TestClientAfterAccountAddBalance(t *testing.T) {
	t.Skip("TODO")
	// 1. update balance in DB
}

func TestClientPreAccountReturnBalance(t *testing.T) {
	t.Skip("TODO")
	// 1. check PSC balance PSC.balanceOf()
	// 2. PSC.returnBalanceERC20()
}

func TestClientAfterAccountReturnBalance(t *testing.T) {
	t.Skip("TODO")
	// 1. update balance in DB
}
