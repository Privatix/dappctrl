package worker

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
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
	fixture.Channel.Client = data.HexFromBytes(clientAddr.Bytes())
	env.updateInTestDB(t, fixture.Channel)

	auth := bind.NewKeyedTransactor(key)
	env.ethBack.SetTransaction(t, auth, nil)

	// Create related eth log record.
	var deposit int64 = 100
	logData, err := logChannelCreatedDataArguments.Pack(big.NewInt(deposit))
	if err != nil {
		t.Fatal(err)
	}
	agentAddr := data.TestToAddress(t, fixture.Account.EthAddr)
	topics := data.LogTopics{
		// First topic is always the event.
		common.BytesToHash([]byte{}),
		common.BytesToHash(agentAddr.Bytes()),
		common.BytesToHash(clientAddr.Bytes()),
		data.TestToHash(t, fixture.Offering.Hash),
	}
	setJobData(t, db, fixture.job, &data.JobData{
		EthLog: &data.JobEthLog{
			TxHash: data.HexFromBytes(env.ethBack.Tx.Hash().Bytes()),
			Data:   logData,
			Topics: topics,
		},
	})

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
	expectedClient := data.HexFromBytes(clientAddr.Bytes())
	if channel.Client != expectedClient {
		t.Fatalf("wanted client addr: %v, got: %v", expectedClient,
			channel.Client)
	}
	expectedAgent := data.HexFromBytes(agentAddr.Bytes())
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
	testAfterChannelTopUp(t, true)
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
	env := newWorkerTest(t)
	fxt := env.newTestFixture(t, data.JobAgentAfterUncooperativeCloseRequest,
		data.JobChannel)
	defer env.close()
	defer fxt.close()

	testChangesStatus := func(svcStatus string, balance uint64) {
		fxt.Channel.ServiceStatus = svcStatus
		fxt.Channel.ReceiptBalance = balance
		env.updateInTestDB(t, fxt.Channel)

		runJob(t, env.worker.AgentAfterUncooperativeCloseRequest,
			fxt.job)

		testChannelStatusChanged(t, fxt.job, env,
			data.ChannelInChallenge)
	}

	// Channel in challenge and s ervice terminate created
	testChangesStatus(data.ServiceSuspended, 0)
	env.deleteJob(t,
		data.JobAgentPreServiceTerminate,
		data.JobChannel,
		fxt.Channel.ID)

	// Terminated channel in challenge and service terminate not created
	testChangesStatus(data.ServiceTerminated, 0)
	env.jobNotCreated(t, fxt.Channel.ID, data.JobAgentPreServiceTerminate)

	testCommonErrors(t, env.worker.AgentAfterUncooperativeCloseRequest,
		*fxt.job)
}

func TestAgentAfterUncooperativeClose(t *testing.T) {
	env := newWorkerTest(t)
	fxt := env.newTestFixture(t, data.JobAgentAfterUncooperativeClose,
		data.JobChannel)
	defer env.close()
	defer fxt.close()

	testStatusChangedAndUpdateBalancesJobCreated := func(svcStatus string) {
		fxt.Channel.ServiceStatus = svcStatus
		env.updateInTestDB(t, fxt.Channel)
		fxt.Offering.CurrentSupply = 0
		env.updateInTestDB(t, fxt.Offering)

		runJob(t, env.worker.AgentAfterUncooperativeClose, fxt.job)

		// Test update balances job was created.
		env.deleteJob(t, data.JobAccountUpdateBalances,
			data.JobAccount, fxt.Account.ID)

		testChannelStatusChanged(t,
			fxt.job,
			env,
			data.ChannelClosedUncoop)
	}

	testStatusChangedAndUpdateBalancesJobCreated(data.ServicePending)
	// Test agent pre service terminate job created.
	env.deleteJob(t, data.JobAgentPreServiceTerminate, data.JobChannel,
		fxt.Channel.ID)

	// testStatusChangedAndUpdateBalancesJobCreated(data.ServiceTerminated)
	// env.jobNotCreated(t, fxt.Channel.ID, data.JobAgentPreServiceTerminate)

	// testCommonErrors(t, env.worker.AgentAfterUncooperativeClose,
	// 	*fxt.job)
}

func TestAgentAfterCooperativeClose(t *testing.T) {
	// set ch_status="closed_coop"
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentAfterCooperativeClose,
		data.JobChannel)
	fixture.Offering.CurrentSupply = 0
	env.updateInTestDB(t, fixture.Offering)
	defer env.close()
	defer fixture.close()

	runJob(t, env.worker.AgentAfterCooperativeClose, fixture.job)

	testChannelStatusChanged(t, fixture.job, env, data.ChannelClosedCoop)

	// Test update balances job was created.
	env.deleteJob(t, data.JobAccountUpdateBalances,
		data.JobAccount, fixture.Account.ID)

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

func testAgentPreServiceTerminate(t *testing.T, receiptBalance uint64) {
	// svc_status="Terminated"
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentPreServiceTerminate,
		data.JobChannel)
	defer env.close()
	defer fixture.close()

	fixture.Channel.ReceiptBalance = receiptBalance
	env.updateInTestDB(t, fixture.Channel)

	runJob(t, env.worker.AgentPreServiceTerminate, fixture.job)

	testServiceStatusChanged(t, fixture.job, env, data.ServiceTerminated)

	if receiptBalance > 0 {
		testCooperativeCloseCalled(t, env, fixture)
	}

	testCommonErrors(t, env.worker.AgentPreServiceTerminate, *fixture.job)
}

func TestAgentPreServiceTerminate(t *testing.T) {
	testAgentPreServiceTerminate(t, 0)
	testAgentPreServiceTerminate(t, 1)
}

func testCooperativeCloseCalled(t *testing.T, env *workerTest,
	fixture *workerTestFixture) {
	// Test eth transaction was recorder.
	defer env.deleteEthTx(t, fixture.job.ID)

	agentAddr := data.TestToAddress(t, fixture.Channel.Agent)

	offeringHash := data.TestToHash(t, fixture.Offering.Hash)

	balance := new(big.Int).SetUint64(fixture.Channel.ReceiptBalance)

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

	env.ethBack.TestCalled(t, "CooperativeClose", agentAddr,
		env.gasConf.PSC.CooperativeClose, agentAddr,
		uint32(fixture.Channel.Block),
		[common.HashLength]byte(offeringHash), balance,
		balanceMsgSig, closingSig)
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
	fxt := env.newTestFixture(t, data.JobAgentPreEndpointMsgCreate,
		data.JobChannel)
	defer env.close()
	defer fxt.close()

	runJob(t, env.worker.AgentPreEndpointMsgCreate, fxt.job)

	endpoint := &data.Endpoint{}
	if err := env.db.SelectOneTo(endpoint,
		"WHERE template=$1 AND channel=$2 AND id<>$3",
		fxt.TemplateAccess.ID,
		fxt.Channel.ID, fxt.Endpoint.ID); err != nil {
		t.Fatalf("could not find %T: %v", endpoint, err)
	}
	env.deleteFromTestDB(t, endpoint)

	if endpoint.RawMsg == "" {
		t.Fatal("raw msg is not set")
	}

	rawMsgBytes := data.TestToBytes(t, endpoint.RawMsg)
	expectedHash := ethcrypto.Keccak256(rawMsgBytes)
	if data.HexFromBytes(expectedHash) != endpoint.Hash {
		t.Fatal("wrong hash stored")
	}

	channel := &data.Channel{}
	env.findTo(t, channel, fxt.Channel.ID)
	if channel.Password == fxt.Channel.Password ||
		channel.Salt == fxt.Channel.Salt {
		t.Fatal("password is not stored in channel")
	}

	testCommonErrors(t, env.worker.AgentPreEndpointMsgCreate, *fxt.job)
}

func TestAgentPreOfferingMsgBCPublish(t *testing.T) {
	// 1. PSC.registerServiceOffering()
	// 2. msg_status="bchain_publishing"
	// 3. offer_status="registered"
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentPreOfferingMsgBCPublish,
		data.JobOffering)
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

	country := "YY"
	fixture.Product.Country = &country
	env.updateInTestDB(t, fixture.Product)

	minDeposit := data.MinDeposit(fixture.Offering)

	env.ethBack.BalancePSC = new(big.Int).SetUint64(minDeposit*
		uint64(fixture.Offering.Supply) + 1)
	env.ethBack.BalanceEth = big.NewInt(99999)

	runJob(t, env.worker.AgentPreOfferingMsgBCPublish, fixture.job)

	agentAddr := data.TestToAddress(t, fixture.Channel.Agent)

	offering := &data.Offering{}
	env.findTo(t, offering, fixture.Offering.ID)

	if offering.Country != *fixture.Product.Country {
		t.Fatalf("expected: %s, got: %s",
			*fixture.Product.Country, offering.Country)
	}

	offeringHash := data.TestToHash(t, offering.Hash)

	env.ethBack.TestCalled(t, "RegisterServiceOffering", agentAddr,
		env.gasConf.PSC.RegisterServiceOffering,
		[common.HashLength]byte(offeringHash),
		new(big.Int).SetUint64(minDeposit), offering.Supply,
		data.OfferingSOMCTor, data.Base64String(""))

	offering = &data.Offering{}
	env.findTo(t, offering, fixture.Offering.ID)
	if offering.Status != data.MsgBChainPublishing {
		t.Fatalf("wrong msg status, wanted: %s, got: %s",
			data.MsgBChainPublishing, offering.Status)
	}
	if offering.OfferStatus != data.OfferRegistering {
		t.Fatalf("wrong offering status, wanted: %s, got: %s",
			data.OfferRegistering, offering.OfferStatus)
	}

	testCommonErrors(t, env.worker.AgentPreOfferingMsgBCPublish,
		*fixture.job)
}

func TestAgentAfterOfferingMsgBCPublish(t *testing.T) {
	// 1. msg_status="bchain_published"
	// 2. "preOfferingMsgSOMCPublish"
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobAgentAfterOfferingMsgBCPublish,
		data.JobOffering)
	defer env.close()
	defer fixture.close()

	logData, err := logOfferingCreatedDataArguments.Pack(uint16(1),
		data.OfferingSOMCTor, "")
	if err != nil {
		t.Fatal(err)
	}

	var blockNumberUpdated uint64 = 100

	setJobData(t, env.db, fixture.job, &data.JobData{
		EthLog: &data.JobEthLog{
			Block: blockNumberUpdated,
			Data:  logData,
			Topics: data.LogTopics{
				common.BytesToHash([]byte{}),
				common.BytesToHash([]byte{}),
				common.BytesToHash([]byte{}),
				common.BytesToHash([]byte{}),
			},
		},
	})

	runJob(t, env.worker.AgentAfterOfferingMsgBCPublish, fixture.job)

	offering := &data.Offering{}
	env.findTo(t, offering, fixture.Offering.ID)
	if offering.Status != data.MsgBChainPublished {
		t.Fatalf("wrong msg status, wanted: %s, got: %s",
			data.MsgBChainPublished, offering.Status)
	}
	if offering.OfferStatus != data.OfferRegistered {
		t.Fatalf("wrong offer status, wanted: %s, got: %s",
			data.OfferRegistered, offering.OfferStatus)
	}

	if offering.BlockNumberUpdated != blockNumberUpdated {
		t.Fatalf("wrong block number updated, wanted: %d, got: %d",
			blockNumberUpdated, offering.BlockNumberUpdated)
	}

	testCommonErrors(t, env.worker.AgentAfterOfferingMsgBCPublish,
		*fixture.job)
}

func TestAgentAfterOfferingDelete(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobAgentAfterOfferingDelete, data.JobOffering)
	defer fxt.close()

	runJob(t, env.worker.AgentAfterOfferingDelete, fxt.job)

	updated := data.Offering{}
	env.findTo(t, &updated, fxt.job.RelatedID)

	if updated.OfferStatus != data.OfferRemoved {
		t.Fatalf("expected offering status: %s, got: %s",
			data.OfferRemoved, updated.OfferStatus)
	}

	testCommonErrors(t, env.worker.AgentAfterOfferingDelete, *fxt.job)
}

func TestAgentPreOfferingDelete(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobAgentPreOfferingDelete, data.JobOffering)
	defer fxt.close()

	setJobData(t, env.db, fxt.job, &data.JobPublishData{
		GasPrice: 123,
	})

	// Offering must be registered.
	if env.worker.AgentPreOfferingDelete(fxt.job) == nil {
		t.Fatal("offering status not validated")
	}

	fxt.Offering.OfferStatus = data.OfferRegistered
	env.updateInTestDB(t, fxt.Offering)

	env.ethBack.OfferingIsActive = true
	removePeriodSetting := &data.Setting{
		Key:   data.SettingsPeriodRemove,
		Value: "10",
		Name:  data.SettingsPeriodRemove,
	}
	env.insertToTestDB(t, removePeriodSetting)
	defer env.deleteFromTestDB(t, removePeriodSetting)
	env.ethBack.OfferUpdateBlockNumber = 10
	env.ethBack.BlockNumber = big.NewInt(10)

	err := env.worker.AgentPreOfferingDelete(fxt.job)
	if err != ErrOfferingDeletePeriodIsNotOver {
		t.Fatal("must check offering delete period")
	}

	env.ethBack.BlockNumber = big.NewInt(100)
	runJob(t, env.worker.AgentPreOfferingDelete, fxt.job)

	// Test transaction was recorded.
	env.deleteEthTx(t, fxt.job.ID)

	agentAddr := data.TestToAddress(t, fxt.Offering.Agent)
	offeringHash := data.TestToHash(t, fxt.Offering.Hash)
	env.ethBack.TestCalled(t, "RemoveServiceOffering", agentAddr,
		env.worker.gasConf.PSC.RemoveServiceOffering,
		[common.HashLength]byte(offeringHash))

	env.db.Reload(fxt.Offering)
	if fxt.Offering.OfferStatus != data.OfferRemoving {
		t.Fatal("offering status not updated")
	}
}

func TestAgentPreOfferingPopUp(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t, data.JobAgentPreOfferingPopUp, data.JobOffering)
	defer fxt.close()

	setJobData(t, env.db, fxt.job, &data.JobPublishData{GasPrice: 123})

	duplicatedJob := *fxt.job
	duplicatedJob.ID = util.NewUUID()

	// Offering must be registered.
	if env.worker.AgentPreOfferingPopUp(fxt.job) == nil {
		t.Fatal("offering status not validated")
	}

	env.insertToTestDB(t, &duplicatedJob)
	defer env.deleteFromTestDB(t, &duplicatedJob)

	fxt.Offering.OfferStatus = data.OfferRegistered
	env.updateInTestDB(t, fxt.Offering)

	if err := env.worker.AgentPreOfferingPopUp(
		fxt.job); err != ErrUncompletedJobsExists {
		t.Fatal("no active jobs")
	}

	duplicatedJob.Status = data.JobDone
	env.updateInTestDB(t, &duplicatedJob)

	if err := env.worker.AgentPreOfferingPopUp(
		fxt.job); err != ErrOfferingNotActive {
		t.Fatal("offering is active")
	}

	env.ethBack.OfferingIsActive = true
	popupPeriodSetting := &data.Setting{
		Key:   data.SettingsPeriodPopUp,
		Value: "3",
		Name:  data.SettingsPeriodPopUp,
	}
	env.insertToTestDB(t, popupPeriodSetting)
	defer env.deleteFromTestDB(t, popupPeriodSetting)
	env.ethBack.OfferUpdateBlockNumber = 3
	env.ethBack.BlockNumber = big.NewInt(4)

	if err := env.worker.AgentPreOfferingPopUp(
		fxt.job); err != ErrPopUpPeriodIsNotOver {
		t.Fatal("period of challenge has expired")
	}

	env.ethBack.BlockNumber = big.NewInt(7)
	runJob(t, env.worker.AgentPreOfferingPopUp, fxt.job)

	// Test transaction was recorded.
	env.deleteEthTx(t, fxt.job.ID)

	agentAddr := data.TestToAddress(t, fxt.Offering.Agent)
	offeringHash := data.TestToHash(t, fxt.Offering.Hash)
	env.ethBack.TestCalled(t, "PopupServiceOffering", agentAddr,
		env.worker.gasConf.PSC.PopupServiceOffering,
		[common.HashLength]byte(offeringHash),
		fxt.Offering.SOMCType, fxt.Offering.SOMCData)

	env.db.Reload(fxt.Offering)

	if fxt.Offering.OfferStatus != data.OfferPoppingUp {
		t.Fatal("offering status not updated")
	}
}

func TestAgentAfterOfferingPopUp(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobAgentAfterOfferingPopUp, data.JobOffering)
	defer fxt.close()

	fxt.job.RelatedID = util.NewUUID()
	env.updateInTestDB(t, fxt.job)

	offeringHash := data.TestToHash(t, fxt.Offering.Hash)
	agentAddr := data.TestToAddress(t, fxt.Account.EthAddr)

	topics := data.LogTopics{
		// First topic is always the event.
		common.BytesToHash([]byte{}),
		common.BytesToHash(agentAddr.Bytes()),
		offeringHash,
		common.BigToHash(big.NewInt(100)),
	}

	logData, err := logOfferingCreatedDataArguments.Pack(
		uint16(1), data.OfferingSOMCTor, "")
	if err != nil {
		t.Fatal(err)
	}

	ethLog := &data.JobEthLog{
		Data:   logData,
		Topics: topics,
		Block:  12345,
	}
	setJobData(t, db, fxt.job, &data.JobData{
		EthLog: ethLog,
	})

	runJob(t, env.worker.AgentAfterOfferingPopUp, fxt.job)

	offering := data.Offering{}
	env.findTo(t, &offering, fxt.Offering.ID)

	if offering.BlockNumberUpdated != ethLog.Block {
		t.Fatal("offering block number was not updated")
	}

	if offering.OfferStatus != data.OfferPoppedUp {
		t.Fatal("offering status not updated")
	}
}
