package worker

import (
	"bytes"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/privatix/dappctrl/somc"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
)

func TestAgentAfterChannelCreate(t *testing.T) {
	// 1. GetTransactionByHash to retrieve public key
	// 2. Derive public Client's public key
	// 3. Add public key to `users` (ingore on duplicate)
	// 4. Add channel to `channels`
	// 5. ch_status="Active"
	// 6. svc_status="Pending"
	// 7. "preEndpointMsgCreate"
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentAfterChannelCreate,
		data.JobChannel)
	defer env.close()
	defer fixture.close()

	key, err := ethcrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	fixture.Channel.Client = data.FromBytes(ethcrypto.PubkeyToAddress(key.PublicKey).Bytes())
	env.updateInTestDB(t, fixture.Channel)

	auth := bind.NewKeyedTransactor(key)
	env.ethBack.setTransaction(t, auth, nil)

	ethLog := data.NewTestEthLog()
	ethLog.TxHash = data.FromBytes(env.ethBack.tx.Hash().Bytes())
	ethLog.Job = fixture.job.ID
	env.insertToTestDB(t, ethLog)
	defer env.deleteFromTestDB(t, ethLog)

	runJob(t, env.worker.AgentAfterChannelCreate, fixture.job)

	channel := &data.Channel{}
	env.findTo(t, channel, fixture.Channel.ID)

	if channel.ChannelStatus != data.ChannelActive {
		t.Fatalf("wanted %s, got: %s", data.ChannelActive,
			channel.ChannelStatus)
	}
	if channel.ServiceStatus != data.ServicePending {
		t.Fatalf("wanted %s, got: %s", data.ServicePending,
			channel.ServiceStatus)
	}

	user := &data.User{}
	if err := env.db.FindOneTo(user, "eth_addr", channel.Client); err != nil {
		t.Fatal(err)
	}
	expected := data.FromBytes(ethcrypto.FromECDSAPub(&key.PublicKey))
	if user.PublicKey != expected {
		t.Fatalf("wanted: %v, got: %v", expected, user.PublicKey)
	}

	// Test pre service create created.
	env.deleteJob(t, data.JobAgentPreEndpointMsgCreate, data.JobChannel, channel.ID)
}

func TestAgentAfterChannelTopUp(t *testing.T) {
	// 1. Add deposit to channels.total_deposit
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentAfterChannelTopUp,
		data.JobChannel)
	defer env.close()
	defer fixture.close()

	block := fixture.Channel.Block
	addedDeposit := big.NewInt(1)

	eventData, err := logChannelTopUpDataArguments.Pack(block, addedDeposit)
	if err != nil {
		t.Fatal(err)
	}

	agentAddr := data.TestToAddress(t, fixture.Channel.Agent)
	clientAddr := data.TestToAddress(t, fixture.Channel.Client)
	offeringHash := data.TestToHash(t, fixture.Offering.Hash)
	topics, err := json.Marshal([]common.Hash{
		common.BytesToHash(agentAddr.Bytes()),
		common.BytesToHash(clientAddr.Bytes()),
		offeringHash,
	})
	if err != nil {
		t.Fatal(err)
	}

	ethLog := data.NewTestEthLog()
	ethLog.Job = fixture.job.ID
	ethLog.Data = data.FromBytes(eventData)
	ethLog.Topics = topics
	env.insertToTestDB(t, ethLog)
	defer env.deleteFromTestDB(t, ethLog)

	runJob(t, env.worker.AgentAfterChannelTopUp, fixture.job)

	channel := &data.Channel{}
	env.findTo(t, channel, fixture.Channel.ID)

	diff := channel.TotalDeposit - fixture.Channel.TotalDeposit
	if diff != addedDeposit.Uint64() {
		t.Fatal("total deposit not updated")
	}

	testCommonErrors(t, env.worker.AgentAfterChannelTopUp, *fixture.job)
}

func testChannelStatusChanged(t *testing.T,
	job *data.Job, env *workerTest, newStatus string) {
	updated := &data.Channel{}
	env.findTo(t, updated, job.RelatedID)

	if newStatus != updated.ChannelStatus {
		t.Fatalf("wanted: %s, got: %s", newStatus, updated.ChannelStatus)
	}
}

func TestAgentAfterUncooperativeCloseRequest(t *testing.T) {
	// 1. set ch_status="in_challenge"
	// 2. if channels.receipt_balance > 0
	//   then "preCooperativeClose"
	//   else "preServiceTerminate"

	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentAfterUncooperativeCloseRequest,
		data.JobChannel)
	defer env.close()
	defer fixture.close()

	testChangesStatusAndCreatesJob := func(t *testing.T, balance uint64, jobType string) {
		fixture.Channel.ReceiptBalance = balance
		env.updateInTestDB(t, fixture.Channel)
		runJob(t, env.worker.AgentAfterUncooperativeCloseRequest,
			fixture.job)
		testChannelStatusChanged(t, fixture.job, env,
			data.ChannelInChallenge)
		env.deleteJob(t,
			jobType,
			data.JobChannel,
			fixture.Channel.ID)
	}

	t.Run("ChannelInChallengeAndServiceTerminateCreated", func(t *testing.T) {
		testChangesStatusAndCreatesJob(t, 0, data.JobAgentPreServiceTerminate)
	})

	t.Run("ChannelInChallengeAndCoopCloseJobCreated", func(t *testing.T) {
		testChangesStatusAndCreatesJob(t, 1, data.JobAgentPreCooperativeClose)
	})

	testCommonErrors(t, env.worker.AgentAfterUncooperativeCloseRequest,
		*fixture.job)
}

func TestAgentAfterUncooperativeClose(t *testing.T) {
	// 1. set ch_status="closed_uncoop"
	// 2. "preServiceTerminate"

	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentAfterUncooperativeClose,
		data.JobChannel)
	defer env.close()
	defer fixture.close()

	runJob(t, env.worker.AgentAfterUncooperativeClose, fixture.job)

	testChannelStatusChanged(t,
		fixture.job,
		env,
		data.ChannelClosedUncoop)

	// Test agent pre service terminate job created.
	env.deleteJob(t, data.JobAgentPreServiceTerminate, data.JobChannel,
		fixture.Channel.ID)

	testCommonErrors(t, env.worker.AgentAfterUncooperativeClose,
		*fixture.job)
}

func TestAgentPreCooperativeClose(t *testing.T) {
	// 1. PSC.cooperativeClose()
	// 2. "preServiceTerminate"
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentPreCooperativeClose,
		data.JobChannel)
	defer env.close()
	defer fixture.close()

	runJob(t, env.worker.AgentPreCooperativeClose, fixture.job)

	agentAddr := data.TestToAddress(t, fixture.Channel.Agent)

	offeringHash := data.TestToHash(t, fixture.Offering.Hash)

	balance := big.NewInt(int64(fixture.Channel.ReceiptBalance))

	balanceMsgSig := data.TestToBytes(t, fixture.Channel.ReceiptSignature)

	clientAddr := data.TestToAddress(t, fixture.Channel.Client)

	balanceHash := eth.BalanceClosingHash(clientAddr, conf.pscAddr,
		uint32(fixture.Channel.Block), offeringHash,
		balance)

	key, err := data.TestToPrivateKey(fixture.Account.PrivateKey, data.TestPassword)
	if err != nil {
		t.Fatal(err)
	}

	closingSig, err := ethcrypto.Sign(balanceHash, key)
	if err != nil {
		t.Fatal(err)
	}

	env.ethBack.testCalled(t, "CooperativeClose", agentAddr, agentAddr,
		uint32(fixture.Channel.Block),
		[common.HashLength]byte(offeringHash), balance,
		balanceMsgSig, closingSig)

	// Test agent pre service terminate job created.
	env.deleteJob(t, data.JobAgentPreServiceTerminate, data.JobChannel, fixture.Channel.ID)

	testCommonErrors(t, env.worker.AgentPreCooperativeClose, *fixture.job)
}

func TestAgentAfterCooperativeClose(t *testing.T) {
	// set ch_status="closed_coop"
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentAfterCooperativeClose,
		data.JobChannel)
	defer env.close()
	defer fixture.close()

	runJob(t, env.worker.AgentAfterCooperativeClose, fixture.job)

	testChannelStatusChanged(t, fixture.job, env, data.ChannelClosedCoop)

	testCommonErrors(t, env.worker.AgentAfterCooperativeClose, *fixture.job)
}

func testServiceStatusChanged(t *testing.T,
	job *data.Job, env *workerTest, newStatus string) {
	updated := &data.Channel{}
	env.findTo(t, updated, job.RelatedID)

	if newStatus != updated.ServiceStatus {
		t.Fatalf("wanted: %s, got: %s", newStatus, updated.ChannelStatus)
	}
}

func TestAgentPreServiceSuspend(t *testing.T) {
	// svc_status="Suspended"
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentPreServiceSuspend,
		data.JobChannel)
	defer env.close()
	defer fixture.close()

	runJob(t, env.worker.AgentPreServiceSuspend, fixture.job)

	testServiceStatusChanged(t, fixture.job, env, data.ServiceSuspended)

	testCommonErrors(t, env.worker.AgentPreServiceSuspend, *fixture.job)
}

func TestAgentPreServiceUnsuspend(t *testing.T) {
	// svc_status="Active"ing.T) {
	// svc_status="Suspended"
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentPreServiceUnsuspend,
		data.JobChannel)
	defer env.close()
	defer fixture.close()

	fixture.Channel.ServiceStatus = data.ServiceSuspended
	env.updateInTestDB(t, fixture.Channel)

	runJob(t, env.worker.AgentPreServiceUnsuspend, fixture.job)

	testServiceStatusChanged(t, fixture.job, env, data.ServiceActive)

	testCommonErrors(t, env.worker.AgentPreServiceUnsuspend, *fixture.job)
}

func TestAgentPreServiceTerminate(t *testing.T) {
	// svc_status="Terminated"
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentPreServiceTerminate,
		data.JobChannel)
	defer env.close()
	defer fixture.close()

	runJob(t, env.worker.AgentPreServiceTerminate, fixture.job)

	testServiceStatusChanged(t, fixture.job, env, data.ServiceTerminated)

	testCommonErrors(t, env.worker.AgentPreServiceTerminate, *fixture.job)
}

func TestAgentPreEndpointMsgCreate(t *testing.T) {
	// generate password
	// store password in DB.channels.password + DB.channels.salt
	// fill & encrypt & sign endpoint message
	// store msg in DB.endpoints filling only "NOT NULL" fields
	// store raw endpoint message in DB.endpoints.raw_msg
	// msg_status="unpublished"
	// "preEndpointMsgSOMCPublish"
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentPreEndpointMsgCreate,
		data.JobChannel)
	defer env.close()
	defer fixture.close()

	runJob(t, env.worker.AgentPreEndpointMsgCreate, fixture.job)

	endpoint := &data.Endpoint{}
	if err := env.db.SelectOneTo(endpoint,
		"where template=$1 and channel=$2 and status=$3",
		fixture.TemplateAccess.ID,
		fixture.Channel.ID, data.MsgUnpublished); err != nil {
		t.Fatal("failed to get desired endpoint: ", err)
	}
	defer env.deleteFromTestDB(t, endpoint)

	if endpoint.RawMsg == "" {
		t.Fatal("raw msg is not set")
	}

	rawMsgBytes := data.TestToBytes(t, endpoint.RawMsg)
	expectedHash := ethcrypto.Keccak256(rawMsgBytes)
	if data.FromBytes(expectedHash) != endpoint.Hash {
		t.Fatal("wrong hash stored")
	}

	channel := &data.Channel{}
	env.findTo(t, channel, fixture.Channel.ID)
	if channel.Password == fixture.Channel.Password ||
		channel.Salt == fixture.Channel.Salt {
		t.Fatal("password is not stored in channel")
	}

	// Check pre publish job created.
	env.deleteJob(t, data.JobAgentPreEndpointMsgSOMCPublish,
		data.JobEndpoint, endpoint.ID)

	testCommonErrors(t, env.worker.AgentPreEndpointMsgCreate, *fixture.job)
}

func TestAgentPreEndpointMsgSOMCPublish(t *testing.T) {
	// 1. publish to SOMC
	// 2. set msg_status="msg_channel_published"
	// 3. "afterEndpointMsgSOMCPublish"
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t,
		data.JobAgentPreEndpointMsgSOMCPublish, data.JobEndpoint)
	defer env.close()
	defer fixture.close()

	somcEndpointChan := make(chan somc.TestEndpointParams)
	go func() {
		somcEndpointChan <- env.fakeSOMC.ReadPublishEndpoint(t)
	}()

	workerF := env.worker.AgentPreEndpointMsgSOMCPublish
	runJob(t, workerF, fixture.job)

	select {
	case ret := <-somcEndpointChan:
		if ret.Channel != fixture.Endpoint.Channel {
			t.Fatal("wrong channel used to publish endpoint")
		}
		msgBytes := data.TestToBytes(t, fixture.Endpoint.RawMsg)
		if !bytes.Equal(msgBytes, ret.Endpoint) {
			t.Fatal("wrong endpoint sent to somc")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timeout")
	}

	endpoint := &data.Endpoint{}
	env.findTo(t, endpoint, fixture.Endpoint.ID)
	if endpoint.Status != data.MsgChPublished {
		t.Fatal("endpoint status is not updated")
	}

	// Test after publish job created.
	env.deleteJob(t, data.JobAgentAfterEndpointMsgSOMCPublish,
		data.JobChannel, endpoint.Channel)

	testCommonErrors(t, workerF, *fixture.job)
}

func testAgentAfterEndpointMsgSOMCPublish(t *testing.T,
	fixture *workerTestFixture, env *workerTest,
	setupPrice uint64, billingType, expectedStatus string) {

	fixture.Channel.ServiceStatus = data.ServicePending
	env.updateInTestDB(t, fixture.Channel)

	fixture.Offering.SetupPrice = setupPrice
	fixture.Offering.BillingType = billingType
	env.updateInTestDB(t, fixture.Offering)

	runJob(t, env.worker.AgentAfterEndpointMsgSOMCPublish, fixture.job)

	testServiceStatusChanged(t, fixture.job, env, expectedStatus)
}

func TestAgentAfterEndpointMsgSOMCPublish(t *testing.T) {
	// 1. If pre_paid OR setup_price > 0, then
	// svc_status="Suspended"
	// else svc_status="Active"
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentAfterEndpointMsgSOMCPublish,
		data.JobChannel)
	defer env.close()
	defer fixture.close()

	testAgentAfterEndpointMsgSOMCPublish(t, fixture, env, 0, data.BillingPrepaid,
		data.ServiceSuspended)
	testAgentAfterEndpointMsgSOMCPublish(t, fixture, env, 1, data.BillingPostpaid,
		data.ServiceSuspended)

	testCommonErrors(t, env.worker.AgentAfterEndpointMsgSOMCPublish,
		*fixture.job)
}

func TestAgentPreOfferingMsgBCPublish(t *testing.T) {
	// 1. PSC.registerServiceOffering()
	// 2. msg_status="bchain_publishing"
	// 3. offer_status="register"
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentPreOfferingMsgBCPublish,
		data.JobOfferring)
	defer env.close()
	defer fixture.close()

	runJob(t, env.worker.AgentPreOfferingMsgBCPublish, fixture.job)

	agentAddr := data.TestToAddress(t, fixture.Channel.Agent)

	offering := fixture.Offering

	offeringHash := data.TestToHash(t, offering.Hash)

	minDeposit := offering.MinUnits*offering.UnitPrice + offering.SetupPrice

	env.ethBack.testCalled(t, "RegisterServiceOffering", agentAddr,
		[common.HashLength]byte(offeringHash),
		big.NewInt(int64(minDeposit)), offering.Supply)

	offering = &data.Offering{}
	env.findTo(t, offering, fixture.Offering.ID)
	if offering.Status != data.MsgBChainPublishing {
		t.Fatalf("wrong msg status, wanted: %s, got: %s",
			data.MsgBChainPublishing, offering.Status)
	}
	if offering.OfferStatus != data.OfferRegister {
		t.Fatalf("wrong offering status, wanted: %s, got: %s",
			data.OfferRegister, offering.OfferStatus)
	}

	testCommonErrors(t, env.worker.AgentPreOfferingMsgBCPublish,
		*fixture.job)
}

func TestAgentAfterOfferingMsgBCPublish(t *testing.T) {
	// 1. msg_status="bchain_published"
	// 2. "preOfferingMsgSOMCPublish"
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentAfterOfferingMsgBCPublish,
		data.JobOfferring)
	defer env.close()
	defer fixture.close()

	runJob(t, env.worker.AgentAfterOfferingMsgBCPublish, fixture.job)

	offering := &data.Offering{}
	env.findTo(t, offering, fixture.Offering.ID)
	if offering.Status != data.MsgBChainPublished {
		t.Fatalf("wrong msg status, wanted: %s, got: %s",
			data.MsgBChainPublished, offering.Status)
	}

	// Test somc publish job created.
	env.deleteJob(t, data.JobAgentPreOfferingMsgSOMCPublish, data.JobOfferring,
		offering.ID)

	testCommonErrors(t, env.worker.AgentAfterOfferingMsgBCPublish,
		*fixture.job)
}

func TestAgentPreOfferingMsgSOMCPublish(t *testing.T) {
	// 1. publish to SOMC
	// 2. set msg_status="msg_channel_published"
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t,
		data.JobAgentPreOfferingMsgSOMCPublish, data.JobOfferring)
	defer env.close()
	defer fixture.close()

	somcOfferingsChan := make(chan somc.TestOfferingParams)
	go func() {
		somcOfferingsChan <- env.fakeSOMC.ReadPublishOfferings(t)
	}()

	workerF := env.worker.AgentPreOfferingMsgSOMCPublish
	runJob(t, workerF, fixture.job)

	offering := &data.Offering{}
	env.findTo(t, offering, fixture.Offering.ID)
	if offering.Status != data.MsgChPublished {
		t.Fatal("offering's status is not updated")
	}

	select {
	case ret := <-somcOfferingsChan:
		if offering.RawMsg == "" {
			t.Fatal("offerring message is not filled")
		}
		if offering.Hash == "" {
			t.Fatal("offering hash is not filled")
		}
		if ret.Data != offering.RawMsg {
			t.Fatal("wrong offering published")
		}
		if ret.Hash != offering.Hash {
			t.Fatal("wrong hash stored")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timeout")
	}

	testCommonErrors(t, workerF, *fixture.job)
}

func TestAgentPreAccountAddBalanceApprove(t *testing.T) {
	// 1. check PTC balance PTC.balanceOf()
	// 2. PTC.approve()
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentPreAccountAddBalanceApprove,
		data.JobAccount)
	defer env.close()
	defer fixture.close()

	var transferAmount int64 = 10

	fixture.setJobData(t, data.JobBalanceData{
		Amount: uint(transferAmount),
	})

	env.ethBack.balancePTC = big.NewInt(transferAmount)

	runJob(t, env.worker.AgentPreAccountAddBalanceApprove, fixture.job)

	agentAddr := data.TestToAddress(t, fixture.Account.EthAddr)

	env.ethBack.testCalled(t, "PTCApprove", agentAddr, conf.pscAddr,
		big.NewInt(transferAmount))

	noCallerAddr := common.BytesToAddress([]byte{})
	env.ethBack.testCalled(t, "PTCBalanceOf", noCallerAddr, agentAddr)

	testCommonErrors(t, env.worker.AgentPreAccountAddBalanceApprove,
		*fixture.job)
}

func TestAgentPreAccountAddBalance(t *testing.T) {
	// 1. PSC.addBalanceERC20()
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentPreAccountAddBalance,
		data.JobAccount)
	defer env.close()
	defer fixture.close()

	var transferAmount int64 = 10

	fixture.setJobData(t, data.JobBalanceData{
		Amount: uint(transferAmount),
	})

	runJob(t, env.worker.AgentPreAccountAddBalance, fixture.job)

	agentAddr := data.TestToAddress(t, fixture.Account.EthAddr)

	env.ethBack.testCalled(t, "PSCAddBalanceERC20", agentAddr,
		big.NewInt(transferAmount))

	testCommonErrors(t, env.worker.AgentPreAccountAddBalance, *fixture.job)
}

func TestAgentAfterAccountAddBalance(t *testing.T) {
	// 1. update balance in DB
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentAfterAccountAddBalance,
		data.JobAccount)
	defer env.close()
	defer fixture.close()

	env.ethBack.balancePTC = big.NewInt(100)
	env.ethBack.balancePSC = big.NewInt(200)

	runJob(t, env.worker.AgentAfterAccountAddBalance, fixture.job)

	account := &data.Account{}
	env.findTo(t, account, fixture.Account.ID)
	if account.PTCBalance != 100 {
		t.Fatalf("wrong ptc balance, wanted: %v, got: %v", 100,
			account.PTCBalance)
	}
	if account.PSCBalance != 200 {
		t.Fatalf("wrong psc balance, wanted: %v, got: %v", 200,
			account.PSCBalance)
	}

	testCommonErrors(t, env.worker.AgentAfterAccountAddBalance, *fixture.job)
}

func TestAgentPreAccountReturnBalance(t *testing.T) {
	// 1. check PSC balance PSC.balanceOf()
	// 2. PSC.returnBalanceERC20()
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentPreAccountReturnBalance,
		data.JobAccount)
	defer env.close()
	defer fixture.close()

	var amount int64 = 10

	fixture.setJobData(t, &data.JobBalanceData{
		Amount: uint(amount),
	})

	env.ethBack.balancePSC = big.NewInt(amount)

	runJob(t, env.worker.AgentPreAccountReturnBalance, fixture.job)

	agentAddr := data.TestToAddress(t, fixture.Account.EthAddr)

	noCallerAddr := common.BytesToAddress([]byte{})
	env.ethBack.testCalled(t, "PSCBalanceOf", noCallerAddr, agentAddr)

	env.ethBack.testCalled(t, "PSCReturnBalanceERC20", agentAddr,
		big.NewInt(amount))

	testCommonErrors(t, env.worker.AgentPreAccountReturnBalance, *fixture.job)
}

func TestAgentAfterAccountReturnBalance(t *testing.T) {
	// 1. update balance in DB
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentAfterAccountReturnBalance,
		data.JobAccount)
	defer env.close()
	defer fixture.close()

	env.ethBack.balancePTC = big.NewInt(100)
	env.ethBack.balancePSC = big.NewInt(200)

	runJob(t, env.worker.AgentAfterAccountReturnBalance, fixture.job)

	account := &data.Account{}
	env.findTo(t, account, fixture.Account.ID)
	if account.PTCBalance != 100 {
		t.Fatalf("wrong ptc balance, wanted: %v, got: %v", 100,
			account.PTCBalance)
	}
	if account.PSCBalance != 200 {
		t.Fatalf("wrong psc balance, wanted: %v, got: %v", 200,
			account.PSCBalance)
	}

	testCommonErrors(t, env.worker.AgentAfterAccountReturnBalance, *fixture.job)
}
