package worker

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/privatix/dappctrl/data"
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

	data.ReloadFromTestDB(t, db, fxt.Channel)
	if fxt.Channel.ServiceStatus != data.ServiceActive {
		t.Fatalf("expected %s service status, but got %s",
			data.ServiceActive, fxt.Channel.ServiceStatus)
	}

	var job data.Job
	err := env.worker.db.SelectOneTo(&job, "WHERE id != $1", fxt.job)
	if err != nil {
		t.Fatalf("no new job created")
	}
	defer data.DeleteFromTestDB(t, db, &job)
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
	t.Skip("TODO")
	// 1. set ch_status="closed_uncoop"
}

func TestClientAfterCooperativeClose(t *testing.T) {
	t.Skip("TODO")
	// 1. set ch_status="closed_coop"
	// 2. "preServiceTerminate"
}

func TestClientPreServiceTerminate(t *testing.T) {
	t.Skip("TODO")
	// 1. svc_status="Terminated"
}

func TestClientPreEndpointMsgSOMCGet(t *testing.T) {
	t.Skip("TODO")
	// 1. Get EndpointMessage from SOMC
	// 2. set msg_status="msg_channel_published"
	// 3. svc_status="Active"
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
