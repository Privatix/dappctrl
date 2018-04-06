package worker

import (
	"testing"
)

func TestClientPreChannelCreate(t *testing.T) {
	// 1. Check sufficient internal balance exists PSC.BalanceOf()
	// 2. Check that available SO supply exists
	// 3. Add channel to `channels` with ch_status="Pending"
	// 5. PSC.createChannel()
	t.Skip("TODO")
	// fixture := newTestFixture(t, data.JobClientPreChannelCreate,
	// 	data.JobOfferring)
	// defer fixture.Close()

	// testEthBack.balancePSC = fixture.Offering

	// runJob(t, testWorker.ClientPreChannelCreate, fixture.job)
}

func TestClientAfterChannelCreate(t *testing.T) {
	t.Skip("TODO")
	// 1. ch_status="Active"
	// 2. svc_status="Pending"
	// 3. "preEndpointMsgSOMCGet"
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
