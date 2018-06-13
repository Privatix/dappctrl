package worker

import (
	"bytes"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/somc"
	"github.com/privatix/dappctrl/util"
)

func TestAgentAfterChannelCreate(t *testing.T) {
	// GetTransactionByHash to retrieve public key
	// Derive public Client's public key
	// Add public key to users (ignore on duplicate)
	// Add new channel to DB.channels with DB.channels.id = DB.jobs.related_id
	// ch_status="Active"
	// svc_status="Pending"
	// "preEndpointMsgCreate"
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentAfterChannelCreate,
		data.JobChannel)
	// Related to id of a channel that needs to be created.
	fixture.job.RelatedID = util.NewUUID()
	env.updateInTestDB(t, fixture.job)
	defer env.close()
	defer fixture.close()

	// Create a key for client and mock transaction return.
	key, err := ethcrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	clientAddr := ethcrypto.PubkeyToAddress(key.PublicKey)
	fixture.Channel.Client = data.FromBytes(clientAddr.Bytes())
	env.updateInTestDB(t, fixture.Channel)

	auth := bind.NewKeyedTransactor(key)
	env.ethBack.setTransaction(t, auth, nil)

	// Create related eth log record.
	var deposit int64 = 100
	logData, err := logChannelCreatedDataArguments.Pack(
		big.NewInt(deposit),
		common.HexToHash("0x12312"))
	if err != nil {
		t.Fatal(err)
	}
	agentAddr := data.TestToAddress(t, fixture.Account.EthAddr)
	topics := data.LogTopics{
		// Don't know what the first topic is, but it appears in real logs.
		common.BytesToHash([]byte{}),
		common.BytesToHash(agentAddr.Bytes()),
		common.BytesToHash(clientAddr.Bytes()),
		data.TestToHash(t, fixture.Offering.Hash),
	}
	if err != nil {
		t.Fatal(err)
	}
	ethLog := data.NewTestEthLog()
	ethLog.TxHash = data.FromBytes(env.ethBack.tx.Hash().Bytes())
	ethLog.JobID = &fixture.job.ID
	ethLog.Data = data.FromBytes(logData)
	ethLog.Topics = topics
	env.insertToTestDB(t, ethLog)
	defer env.deleteFromTestDB(t, ethLog)

	runJob(t, env.worker.AgentAfterChannelCreate, fixture.job)

	// Test channel was created.
	channel := &data.Channel{}
	env.findTo(t, channel, fixture.job.RelatedID)
	defer env.deleteFromTestDB(t, channel)

	if channel.ChannelStatus != data.ChannelActive {
		t.Fatalf("wanted %s, got: %s", data.ChannelActive,
			channel.ChannelStatus)
	}
	if channel.ServiceStatus != data.ServicePending {
		t.Fatalf("wanted %s, got: %s", data.ServicePending,
			channel.ServiceStatus)
	}
	expectedClient := data.FromBytes(clientAddr.Bytes())
	if channel.Client != expectedClient {
		t.Fatalf("wanted client addr: %v, got: %v", expectedClient,
			channel.Client)
	}
	expectedAgent := data.FromBytes(agentAddr.Bytes())
	if channel.Agent != expectedAgent {
		t.Fatalf("wanted agent addr: %v, got: %v", expectedAgent,
			channel.Agent)
	}
	if channel.Offering != fixture.Offering.ID {
		t.Fatalf("wanted offering: %s, got: %s", fixture.Offering.ID,
			channel.Offering)
	}
	if channel.TotalDeposit != uint64(deposit) {
		t.Fatalf("wanted total deposit: %v, got: %v", deposit,
			channel.TotalDeposit)
	}

	user := &data.User{}
	if err := env.db.FindOneTo(user, "eth_addr", channel.Client); err != nil {
		t.Fatal(err)
	}
	defer env.deleteFromTestDB(t, user)

	expected := data.FromBytes(ethcrypto.FromECDSAPub(&key.PublicKey))
	if user.PublicKey != expected {
		t.Fatalf("wanted: %v, got: %v", expected, user.PublicKey)
	}

	// Test pre service create created.
	env.deleteJob(t, data.JobAgentPreEndpointMsgCreate, data.JobChannel, channel.ID)
}

func TestAgentAfterChannelTopUp(t *testing.T) {
	// Add deposit to channels.total_deposit
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
	topics := data.LogTopics{
		common.BytesToHash(agentAddr.Bytes()),
		common.BytesToHash(clientAddr.Bytes()),
		offeringHash,
	}
	if err != nil {
		t.Fatal(err)
	}

	ethLog := data.NewTestEthLog()
	ethLog.JobID = &fixture.job.ID
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
	// set ch_status="in_challenge"
	// if channels.receipt_balance > 0
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

	// Test eth transaction was recorder.
	defer env.deleteEthTx(t, fixture.job.ID)

	runJob(t, env.worker.AgentPreCooperativeClose, fixture.job)

	agentAddr := data.TestToAddress(t, fixture.Channel.Agent)

	offeringHash := data.TestToHash(t, fixture.Offering.Hash)

	balance := big.NewInt(int64(fixture.Channel.ReceiptBalance))

	balanceMsgSig := data.TestToBytes(t, *fixture.Channel.ReceiptSignature)

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

	env.ethBack.testCalled(t, "CooperativeClose", agentAddr,
		env.gasConf.PSC.CooperativeClose, agentAddr,
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
	// svc_status="Active"
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
		t.Fatalf("could not find %T: %v", endpoint, err)
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
	case <-time.After(conf.JobHanlderTest.SOMCTimeout * time.Second):
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

	// Test ethTx was recorder.
	defer env.deleteEthTx(t, fixture.job.ID)

	jobData := &data.JobPublishData{GasPrice: 10}
	jobDataB, err := json.Marshal(jobData)
	if err != nil {
		t.Fatal(err)
	}
	fixture.job.Data = jobDataB
	env.updateInTestDB(t, fixture.job)

	minDeposit := fixture.Offering.MinUnits*fixture.Offering.UnitPrice +
		fixture.Offering.SetupPrice

	env.ethBack.balancePSC = big.NewInt(int64(minDeposit*
		uint64(fixture.Offering.Supply) + 1))
	env.ethBack.balanceEth = big.NewInt(99999)

	runJob(t, env.worker.AgentPreOfferingMsgBCPublish, fixture.job)

	agentAddr := data.TestToAddress(t, fixture.Channel.Agent)

	offering := &data.Offering{}
	env.findTo(t, offering, fixture.Offering.ID)

	if offering.RawMsg == fixture.Offering.RawMsg {
		t.Fatal("raw msg was not set")
	}

	if offering.Hash == fixture.Offering.Hash {
		t.Fatal("hash was not set")
	}

	offeringHash := data.TestToHash(t, offering.Hash)

	env.ethBack.testCalled(t, "RegisterServiceOffering", agentAddr,
		env.gasConf.PSC.RegisterServiceOffering,
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

	env.setOfferingHash(t, fixture)
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
		if ret.Data != offering.RawMsg {
			t.Fatal("wrong offering published")
		}
		if ret.Hash != offering.Hash {
			t.Fatal("wrong hash stored")
		}
	case <-time.After(conf.JobHanlderTest.SOMCTimeout * time.Second):
		t.Fatal("timeout")
	}

	testCommonErrors(t, workerF, *fixture.job)
}
