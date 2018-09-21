package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/client/svcrun"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/messages"
	"github.com/privatix/dappctrl/messages/ept"
	"github.com/privatix/dappctrl/messages/offer"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util"
)

func TestClientPreChannelCreate(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientPreChannelCreate, data.JobChannel)
	defer fxt.Close()

	offeringMsg := offer.OfferingMessage(fxt.Account, fxt.TemplateOffer,
		fxt.Offering)
	offeringMsgBytes, _ := json.Marshal(offeringMsg)
	key, _ := data.TestToPrivateKey(fxt.Account.PrivateKey, data.TestPassword)
	packed, _ := messages.PackWithSignature(offeringMsgBytes, key)
	fxt.Offering.RawMsg = data.FromBytes(packed)
	env.updateInTestDB(t, fxt.Offering)

	fxt.job.RelatedType = data.JobChannel
	fxt.job.RelatedID = util.NewUUID()

	setJobData(t, fxt.DB, fxt.job, ClientPreChannelCreateData{
		Account:  fxt.Account.ID,
		Offering: fxt.Offering.ID,
	})

	minDeposit := data.MinDeposit(fxt.Offering)
	env.ethBack.balancePSC = new(big.Int).SetUint64(minDeposit - 1)
	util.TestExpectResult(t, "Job run", ErrInsufficientPSCBalance,
		env.worker.ClientPreChannelCreate(fxt.job))

	env.ethBack.balancePSC = new(big.Int).SetUint64(minDeposit)
	util.TestExpectResult(t, "Job run", ErrOfferingNoSupply,
		env.worker.ClientPreChannelCreate(fxt.job))

	issued := time.Now()
	env.ethBack.offerCurrentSupply = 1

	customDeposit := 99
	env.ethBack.balancePSC = big.NewInt(100)
	setJobData(t, fxt.DB, fxt.job, ClientPreChannelCreateData{
		Account:  fxt.Account.ID,
		Offering: fxt.Offering.ID,
		Deposit:  99,
	})

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
		ch.TotalDeposit != uint64(customDeposit) {
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

	var agentUserRec data.User
	env.selectOneTo(t, &agentUserRec, "WHERE eth_addr=$1 and public_key=$2",
		fxt.Account.EthAddr, fxt.Account.PublicKey)
	env.deleteFromTestDB(t, &agentUserRec)

	env.ethBack.testCalled(t,
		"PSCCreateChannel",
		data.TestToAddress(t, fxt.Account.EthAddr),
		env.gasConf.PSC.CreateChannel,
		data.TestToAddress(t, fxt.Account.EthAddr),
		[common.HashLength]byte(data.TestToHash(t, fxt.Offering.Hash)),
		big.NewInt(int64(customDeposit)),
	)
}

func TestClientAfterChannelCreate(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientAfterChannelCreate, data.JobChannel)
	defer fxt.Close()

	ethLog := data.NewTestEthLog()
	ethLog.JobID = &fxt.job.ID
	data.SaveToTestDB(t, db, ethLog)

	fxt.Channel.ServiceStatus = data.ServicePending
	data.SaveToTestDB(t, db, fxt.Channel)
	defer data.DeleteFromTestDB(t, db, ethLog)

	channelKey, _ := data.ChannelKey(fxt.Channel.Client, fxt.Channel.Agent,
		uint32(ethLog.BlockNumber), fxt.Offering.Hash)

	go func() {
		// Mock reply from SOMC.
		time.Sleep(conf.JobHandlerTest.ReactionDelay * time.Millisecond)
		env.fakeSOMC.WriteGetEndpoint(t, data.FromBytes(channelKey), nil)
	}()

	runJob(t, env.worker.ClientAfterChannelCreate, fxt.job)

	// Test account balance update job was created.
	env.deleteJob(t, data.JobAccountUpdateBalances, data.JobAccount, fxt.UserAcc.ID)

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

	key, err := env.worker.key(env.worker.logger, fxt.Account.PrivateKey)
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

	setJobData(t, fxt.DB, fxt.job, sealed)
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
		data.JobClientAfterEndpointMsgSOMCGet, data.JobChannel)
	defer fxt.Close()

	fxt.job.Data = []byte("\"" + fxt.Endpoint.ID + "\"")

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
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientPreChannelTopUp, data.JobChannel)
	defer fxt.Close()

	setJobData(t, fxt.DB, fxt.job, data.JobPublishData{
		GasPrice: uint64(testTXGasPrice),
	})

	minDeposit := fxt.Offering.UnitPrice*fxt.Offering.MinUnits +
		fxt.Offering.SetupPrice

	env.ethBack.balancePSC = new(big.Int).SetUint64(minDeposit - 1)
	util.TestExpectResult(t, "Job run", ErrInsufficientPSCBalance,
		env.worker.ClientPreChannelTopUp(fxt.job))

	issued := time.Now()
	env.ethBack.balancePSC = new(big.Int).SetUint64(minDeposit)

	runJob(t, env.worker.ClientPreChannelTopUp, fxt.job)

	var tx data.EthTx
	env.selectOneTo(t, &tx, "WHERE related_id = $1", fxt.Channel.ID)
	defer env.deleteFromTestDB(t, &tx)

	if tx.Method != "TopUpChannel" ||
		tx.Status != data.TxSent ||
		tx.JobID == nil || *tx.JobID != fxt.job.ID ||
		tx.Issued.Before(issued) || tx.Issued.After(time.Now()) ||
		tx.AddrFrom != fxt.UserAcc.EthAddr ||
		tx.AddrTo != data.HexFromBytes(env.worker.pscAddr.Bytes()) ||
		tx.Nonce == nil || *tx.Nonce != fmt.Sprint(testTXNonce) ||
		tx.GasPrice != uint64(testTXGasPrice) ||
		tx.Gas != uint64(testTXGasLimit) ||
		tx.RelatedType != data.JobChannel ||
		tx.RelatedID != fxt.Channel.ID {
		t.Fatalf("wrong transaction content")
	}
}

func TestClientAfterChannelTopUp(t *testing.T) {
	testAfterChannelTopUp(t, false)
}

func checkChanStatus(t *testing.T, env *workerTest, channel string,
	status string) {
	var ch data.Channel
	env.findTo(t, &ch, channel)

	if ch.ChannelStatus != status {
		t.Fatal("channel status is wrong")
	}
}

func TestClientPreUncooperativeCloseRequest(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientPreUncooperativeCloseRequest, data.JobChannel)
	defer fxt.Close()

	fxt.Channel.TotalDeposit = 1
	fxt.Channel.ReceiptBalance = 1
	fxt.Channel.ServiceStatus = data.ServiceTerminated
	env.updateInTestDB(t, fxt.Channel)

	issued := time.Now()

	setJobData(t, fxt.DB, fxt.job, data.JobPublishData{
		GasPrice: uint64(testTXGasPrice),
	})

	runJob(t, env.worker.ClientPreUncooperativeCloseRequest, fxt.job)

	checkChanStatus(t, env, fxt.Channel.ID,
		data.ChannelWaitChallenge)

	var tx data.EthTx

	env.selectOneTo(t, &tx, "WHERE related_id = $1", fxt.Channel.ID)
	defer env.deleteFromTestDB(t, &tx)

	if tx.Method != "UncooperativeClose" ||
		tx.Status != data.TxSent ||
		tx.JobID == nil || *tx.JobID != fxt.job.ID ||
		tx.Issued.Before(issued) || tx.Issued.After(time.Now()) ||
		tx.AddrFrom != fxt.Channel.Client ||
		tx.AddrTo != data.HexFromBytes(env.worker.pscAddr.Bytes()) ||
		tx.Nonce == nil || *tx.Nonce != fmt.Sprint(testTXNonce) ||
		tx.GasPrice != uint64(testTXGasPrice) ||
		tx.Gas != uint64(testTXGasLimit) ||
		tx.RelatedType != data.JobChannel ||
		tx.RelatedID != fxt.Channel.ID {
		t.Fatalf("wrong transaction content")
	}

	badChStatus := []string{data.ChannelClosedCoop,
		data.ChannelClosedUncoop, data.ChannelWaitUncoop,
		data.ChannelWaitCoop, data.ChannelWaitChallenge,
		data.ChannelInChallenge}

	badServiceStatus := []string{data.ServiceActive, data.ServiceSuspended}

	fxt.Channel.ReceiptBalance = 2
	env.updateInTestDB(t, fxt.Channel)

	util.TestExpectResult(t, "Job run", ErrChannelReceiptBalance,
		env.worker.ClientPreUncooperativeCloseRequest(fxt.job))

	for _, status := range badServiceStatus {
		fxt.Channel.ServiceStatus = status
		env.updateInTestDB(t, fxt.Channel)

		util.TestExpectResult(t, "Job run", ErrInvalidServiceStatus,
			env.worker.ClientPreUncooperativeCloseRequest(fxt.job))
	}

	for _, status := range badChStatus {
		fxt.Channel.ChannelStatus = status
		env.updateInTestDB(t, fxt.Channel)

		util.TestExpectResult(t, "Job run", ErrInvalidChannelStatus,
			env.worker.ClientPreUncooperativeCloseRequest(fxt.job))
	}

	testCommonErrors(
		t, env.worker.ClientPreUncooperativeCloseRequest, *fxt.job)
}

func TestClientAfterUncooperativeCloseRequest(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientAfterUncooperativeCloseRequest, data.JobChannel)
	defer fxt.Close()

	st := data.Setting{Key: data.SettingEthChallengePeriod, Value: "0"}
	data.InsertToTestDB(t, db, &st)
	defer data.DeleteFromTestDB(t, db, &st)

	runJob(t, env.worker.ClientAfterUncooperativeCloseRequest, fxt.job)

	ch := new(data.Channel)
	env.findTo(t, ch, fxt.Channel.ID)

	if ch.ChannelStatus != data.ChannelInChallenge {
		t.Fatal("channel status is wrong")
	}

	j, err := env.db.FindAllFrom(data.JobTable, "related_id",
		fxt.Channel.ID)

	if err != nil {
		t.Fatal(err)
	}

	var jobs []*data.Job

	for _, v := range j {
		if job, ok := v.(*data.Job); ok {
			jobs = append(jobs, job)
		}
	}

	if len(jobs) != 2 {
		t.Fatal("not all jobs are in the database")
	}
}

func TestClientPreUncooperativeClose(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientPreUncooperativeClose, data.JobChannel)
	defer fxt.Close()

	issued := time.Now()

	runJob(t, env.worker.ClientPreUncooperativeClose, fxt.job)

	var ch data.Channel
	env.findTo(t, &ch, fxt.Channel.ID)
	if ch.ChannelStatus != data.ChannelWaitUncoop {
		t.Fatal("bad channel status")
	}

	var tx data.EthTx
	env.selectOneTo(t, &tx, "WHERE related_id = $1", fxt.Channel.ID)
	defer env.deleteEthTx(t, fxt.job.ID)

	if tx.Method != "Settle" ||
		tx.Status != data.TxSent ||
		tx.JobID == nil || *tx.JobID != fxt.job.ID ||
		tx.Issued.Before(issued) || tx.Issued.After(time.Now()) ||
		tx.AddrFrom != fxt.UserAcc.EthAddr ||
		tx.AddrTo != data.HexFromBytes(env.worker.pscAddr.Bytes()) ||
		tx.Nonce == nil || *tx.Nonce != fmt.Sprint(testTXNonce) ||
		tx.GasPrice != uint64(testTXGasPrice) ||
		tx.Gas != uint64(testTXGasLimit) ||
		tx.RelatedType != data.JobChannel ||
		tx.RelatedID != fxt.Channel.ID {
		t.Fatalf("wrong transaction content")
	}
}

func TestClientAfterUncooperativeClose(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientAfterUncooperativeClose, data.JobChannel)
	defer fxt.Close()

	fxt.Channel.ServiceStatus = data.ServiceTerminated
	fxt.Channel.ChannelStatus = data.ChannelWaitUncoop
	env.updateInTestDB(t, fxt.Channel)

	runJob(t, env.worker.ClientAfterUncooperativeClose, fxt.job)

	// Test update balances job was created.
	env.deleteJob(t,
		data.JobAccountUpdateBalances, data.JobAccount, fxt.Account.ID)

	var ch data.Channel
	env.findTo(t, &ch, fxt.Channel.ID)

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

	testJobCreatedAndStatusChanged := func(svcStatus string) {
		fxt.Channel.ServiceStatus = svcStatus
		fxt.Channel.ChannelStatus = data.ChannelWaitCoop
		env.updateInTestDB(t, fxt.Channel)

		runJob(t, env.worker.ClientAfterCooperativeClose, fxt.job)

		// Test update balances job was created.
		env.deleteJob(t,
			data.JobAccountUpdateBalances, data.JobAccount, fxt.Account.ID)

		var ch data.Channel
		env.findTo(t, &ch, fxt.Channel.ID)

		if ch.ChannelStatus != data.ChannelClosedCoop {
			t.Fatalf("expected %s service status, but got %s",
				data.ChannelClosedCoop, fxt.Channel.ChannelStatus)
		}
	}

	testJobCreatedAndStatusChanged(data.ServicePending)
	// Test terminate job created.
	env.deleteJob(t, data.JobClientPreServiceTerminate, data.JobChannel,
		fxt.Channel.ID)

	testJobCreatedAndStatusChanged(data.ServiceTerminated)
	env.jobNotCreated(t, fxt.Channel.ID, data.JobClientPreServiceTerminate)
}

func runJobCheckingRunnerCall(t *testing.T, env *workerTest,
	workerF func(*data.Job) error, job *data.Job, runnerMethod int) {
	calledMethod := -1
	var calledChannel string
	env.worker.runner = svcrun.Mock(
		func(method int, channel string) (bool, error) {
			calledMethod, calledChannel = method, channel
			return false, nil
		})

	runJob(t, workerF, job)

	if calledMethod != runnerMethod || calledChannel != job.RelatedID {
		t.Fatalf("unexpected service runner call arguments")
	}
}

func TestClientPreServiceTerminate(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientPreServiceTerminate, data.JobChannel)
	defer fxt.Close()

	fxt.Channel.ServiceStatus = data.ServiceActive
	env.updateInTestDB(t, fxt.Channel)

	if _, err := env.worker.processor.TerminateChannel(fxt.Channel.ID,
		data.JobTask, false); err != proc.ErrSameJobExists {
		t.Fatal("job allready exists")
	}

	env.deleteFromTestDB(t, fxt.job)

	jobID, err := env.worker.processor.TerminateChannel(fxt.Channel.ID,
		data.JobTask, false)
	if err != nil {
		t.Fatal(err)
	}

	var job data.Job
	env.findTo(t, &job, jobID)

	runJobCheckingRunnerCall(t, env,
		env.worker.ClientPreServiceTerminate, &job, svcrun.MockStop)

	var ch data.Channel
	env.findTo(t, &ch, fxt.Channel.ID)

	if ch.ServiceStatus != data.ServiceTerminated {
		t.Fatalf("expected %s service status, but got %s",
			data.ServiceTerminated, fxt.Channel.ServiceStatus)
	}
}

func TestClientPreServiceSuspend(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientPreServiceSuspend, data.JobChannel)
	defer fxt.Close()

	fxt.Channel.ServiceStatus = data.ServiceActive
	env.updateInTestDB(t, fxt.Channel)

	if _, err := env.worker.processor.SuspendChannel(fxt.Channel.ID,
		data.JobTask, false); err != proc.ErrActiveJobsExist {
		t.Fatal("job allready exists")
	}

	badStatus := []string{data.ServicePending, data.ServiceSuspended,
		data.ServiceTerminated}

	for _, status := range badStatus {
		fxt.Channel.ServiceStatus = status
		env.updateInTestDB(t, fxt.Channel)
		if _, err := env.worker.processor.SuspendChannel(fxt.Channel.ID,
			data.JobTask, false); err != proc.ErrBadServiceStatus {
			t.Fatal("bad service status")
		}
	}

	fxt.Channel.ServiceStatus = data.ServiceActive
	env.updateInTestDB(t, fxt.Channel)
	env.deleteFromTestDB(t, fxt.job)

	jobID, err := env.worker.processor.SuspendChannel(fxt.Channel.ID,
		data.JobTask, false)
	if err != nil {
		t.Fatal(err)
	}

	var job data.Job
	env.findTo(t, &job, jobID)
	defer env.deleteFromTestDB(t, &job)

	runJobCheckingRunnerCall(t, env,
		env.worker.ClientPreServiceSuspend, &job, svcrun.MockStop)

	var ch data.Channel
	env.findTo(t, &ch, fxt.Channel.ID)

	if ch.ServiceStatus != data.ServiceSuspended {
		t.Fatalf("expected %s service status, but got %s",
			data.ServiceSuspended, fxt.Channel.ServiceStatus)
	}
}

func TestClientPreServiceUnsuspend(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientPreServiceUnsuspend, data.JobChannel)
	defer fxt.Close()

	fxt.Channel.ServiceStatus = data.ServicePending
	env.updateInTestDB(t, fxt.Channel)

	if _, err := env.worker.processor.ActivateChannel(fxt.Channel.ID,
		data.JobTask, false); err != proc.ErrActiveJobsExist {
		t.Fatal("job allready exists")
	}

	badStatus := []string{data.ServiceActive, data.ServiceTerminated}

	for _, status := range badStatus {
		fxt.Channel.ServiceStatus = status
		env.updateInTestDB(t, fxt.Channel)
		if _, err := env.worker.processor.ActivateChannel(fxt.Channel.ID,
			data.JobTask, false); err != proc.ErrBadServiceStatus {
			t.Fatal("bad service status")
		}
	}

	fxt.Channel.ServiceStatus = data.ServicePending
	env.updateInTestDB(t, fxt.Channel)
	env.deleteFromTestDB(t, fxt.job)

	jobID, err := env.worker.processor.ActivateChannel(fxt.Channel.ID,
		data.JobTask, false)
	if err != nil {
		t.Fatal(err)
	}

	var job data.Job
	env.findTo(t, &job, jobID)
	defer env.deleteFromTestDB(t, &job)

	runJobCheckingRunnerCall(t, env,
		env.worker.ClientPreServiceUnsuspend, &job, svcrun.MockStart)

	var ch data.Channel
	env.findTo(t, &ch, fxt.Channel.ID)

	if ch.ServiceStatus != data.ServiceActive {
		t.Fatalf("expected %s service status, but got %s",
			data.ServiceActive, fxt.Channel.ServiceStatus)
	}
}

func TestClientAfterOfferingMsgBCPublish(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientAfterOfferingMsgBCPublish, data.JobOffering)
	defer fxt.Close()

	// Set id for offerring that is about to be created.
	fxt.job.RelatedID = util.NewUUID()
	env.updateInTestDB(t, fxt.job)

	var offeringHash common.Hash

	// Create expected offering.
	expectedOffering := *fxt.Offering
	expectedOffering.ID = fxt.job.RelatedID
	expectedOffering.Status = data.MsgChPublished
	expectedOffering.OfferStatus = data.OfferRegister
	expectedOffering.Country = "US"
	expectedOffering.MinUnits = 100
	msg := offer.OfferingMessage(fxt.Account,
		fxt.TemplateOffer, &expectedOffering)
	msgBytes, err := json.Marshal(msg)
	util.TestExpectResult(t, "Marshall msg", nil, err)
	key, err := env.worker.key(env.worker.logger, fxt.Account.PrivateKey)
	util.TestExpectResult(t, "Get key", nil, err)
	packed, err := messages.PackWithSignature(msgBytes, key)
	util.TestExpectResult(t, "PackWithSignature", nil, err)
	expectedOffering.RawMsg = data.FromBytes(packed)
	offeringHash = common.BytesToHash(crypto.Keccak256(packed))
	expectedOffering.Hash = data.FromBytes(offeringHash.Bytes())

	env.ethBack.offerCurrentSupply = expectedOffering.CurrentSupply
	env.ethBack.offerMaxSupply = expectedOffering.Supply
	env.ethBack.offerMinDeposit = new(big.Int).SetUint64(
		data.MinDeposit(&expectedOffering))

	// Create eth log records used by job.
	var curSupply uint16 = expectedOffering.Supply
	logData, err := logOfferingCreatedDataArguments.Pack(curSupply)
	if err != nil {
		t.Fatal(err)
	}
	agentAddr := data.TestToAddress(t, fxt.Account.EthAddr)
	minDeposit := big.NewInt(10000)
	topics := data.LogTopics{
		common.BytesToHash([]byte{}),
		common.BytesToHash(agentAddr.Bytes()),
		offeringHash,
		common.BytesToHash(minDeposit.Bytes()),
	}
	ethLog := data.NewTestEthLog()
	ethLog.JobID = &fxt.job.ID
	ethLog.Data = data.FromBytes(logData)
	ethLog.Topics = topics
	env.insertToTestDB(t, ethLog)
	defer env.deleteFromTestDB(t, ethLog)

	go func() {
		// Mock reply from SOMC.
		time.Sleep(conf.JobHandlerTest.ReactionDelay * time.Millisecond)
		env.fakeSOMC.WriteFindOfferings(t,
			[]string{expectedOffering.Hash}, [][]byte{packed})
	}()

	runJob(t, env.worker.ClientAfterOfferingMsgBCPublish, fxt.job)

	created := &data.Offering{}
	env.selectOneTo(t, created, "WHERE hash=$1", expectedOffering.Hash)
	defer env.deleteFromTestDB(t, created)

	if expectedOffering.Template != created.Template ||
		expectedOffering.Product != created.Product ||
		created.Status != data.MsgChPublished ||
		expectedOffering.Agent != created.Agent ||
		expectedOffering.RawMsg != created.RawMsg ||
		fxt.Product.Name != created.ServiceName ||
		expectedOffering.Description != created.Description ||
		expectedOffering.Country != created.Country ||
		expectedOffering.Supply != created.Supply ||
		expectedOffering.UnitName != created.UnitName ||
		expectedOffering.BillingType != created.BillingType ||
		expectedOffering.SetupPrice != created.SetupPrice ||
		expectedOffering.UnitPrice != created.UnitPrice ||
		expectedOffering.MinUnits != created.MinUnits ||
		expectedOffering.BillingInterval != created.BillingInterval ||
		expectedOffering.MaxBillingUnitLag != created.MaxBillingUnitLag ||
		expectedOffering.MaxSuspendTime != created.MaxSuspendTime ||
		expectedOffering.MaxInactiveTimeSec != created.MaxInactiveTimeSec ||
		expectedOffering.FreeUnits != created.FreeUnits ||
		!bytes.Equal(expectedOffering.AdditionalParams, created.AdditionalParams) {
		t.Fatal("wrong offering created")
	}
}

func TestClientAfterOfferingDelete(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientAfterOfferingDelete, data.JobOffering)
	defer fxt.Close()

	runJob(t, env.worker.ClientAfterOfferingDelete, fxt.job)

	updated := data.Offering{}
	env.findTo(t, &updated, fxt.job.RelatedID)

	if updated.OfferStatus != data.OfferRemove {
		t.Fatalf("expected offering status: %s, got: %s",
			data.OfferRemove, updated.OfferStatus)
	}

	testCommonErrors(t, env.worker.ClientAfterOfferingDelete, *fxt.job)
}

func TestClientAfterOfferingPopUp(t *testing.T) {
	testClientAfterExistingOfferingPopUp(t)
	testClientAfterNewOfferingPopUp(t)
}

func testClientAfterExistingOfferingPopUp(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientAfterOfferingPopUp, data.JobOffering)
	defer fxt.close()

	fxt.job.RelatedID = util.NewUUID()
	env.updateInTestDB(t, fxt.job)

	offeringHash := data.TestToHash(t, fxt.Offering.Hash)

	topics := data.LogTopics{
		common.BytesToHash([]byte{}),
		common.BytesToHash([]byte{}),
		offeringHash,
	}

	ethLog := data.NewTestEthLog()
	ethLog.JobID = &fxt.job.ID
	ethLog.Topics = topics
	env.insertToTestDB(t, ethLog)
	defer env.deleteFromTestDB(t, ethLog)

	runJob(t, env.worker.ClientAfterOfferingPopUp, fxt.job)

	offering := data.Offering{}
	env.findTo(t, &offering, fxt.Offering.ID)

	if offering.BlockNumberUpdated != ethLog.BlockNumber {
		t.Fatal("offering block number was not updated")
	}
}

func testClientAfterNewOfferingPopUp(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientAfterOfferingPopUp, data.JobOffering)
	defer fxt.close()

	fxt.job.RelatedID = util.NewUUID()
	env.updateInTestDB(t, fxt.job)

	var offeringHash common.Hash

	// Create expected offering.
	expectedOffering := *fxt.Offering
	expectedOffering.ID = fxt.job.RelatedID
	expectedOffering.Status = data.MsgChPublished
	expectedOffering.OfferStatus = data.OfferRegister
	expectedOffering.Country = "US"
	expectedOffering.MinUnits = 100
	msg := offer.OfferingMessage(fxt.Account,
		fxt.TemplateOffer, &expectedOffering)
	msgBytes, err := json.Marshal(msg)
	util.TestExpectResult(t, "Marshall msg", nil, err)
	key, err := env.worker.key(env.worker.logger, fxt.Account.PrivateKey)
	util.TestExpectResult(t, "Get key", nil, err)
	packed, err := messages.PackWithSignature(msgBytes, key)
	util.TestExpectResult(t, "PackWithSignature", nil, err)
	expectedOffering.RawMsg = data.FromBytes(packed)
	offeringHash = common.BytesToHash(crypto.Keccak256(packed))
	expectedOffering.Hash = data.FromBytes(offeringHash.Bytes())

	env.ethBack.offerCurrentSupply = expectedOffering.CurrentSupply
	env.ethBack.offerMaxSupply = expectedOffering.Supply
	env.ethBack.offerMinDeposit = new(big.Int).SetUint64(
		data.MinDeposit(&expectedOffering))

	// Create eth log records used by job.
	agentAddr := data.TestToAddress(t, fxt.Account.EthAddr)
	topics := data.LogTopics{
		common.BytesToHash([]byte{}),
		common.BytesToHash(agentAddr.Bytes()),
		offeringHash,
	}
	ethLog := data.NewTestEthLog()
	ethLog.JobID = &fxt.job.ID
	ethLog.Topics = topics
	env.insertToTestDB(t, ethLog)
	defer env.deleteFromTestDB(t, ethLog)

	go func() {
		// Mock reply from SOMC.
		time.Sleep(conf.JobHandlerTest.ReactionDelay * time.Millisecond)
		env.fakeSOMC.WriteFindOfferings(t,
			[]string{expectedOffering.Hash}, [][]byte{packed})
	}()

	runJob(t, env.worker.ClientAfterOfferingPopUp, fxt.job)

	created := &data.Offering{}
	env.selectOneTo(t, created, "WHERE hash=$1", expectedOffering.Hash)
	defer env.deleteFromTestDB(t, created)

	if expectedOffering.Template != created.Template ||
		expectedOffering.Product != created.Product ||
		created.Status != data.MsgChPublished ||
		expectedOffering.Agent != created.Agent ||
		expectedOffering.RawMsg != created.RawMsg ||
		fxt.Product.Name != created.ServiceName ||
		expectedOffering.Description != created.Description ||
		expectedOffering.Country != created.Country ||
		expectedOffering.Supply != created.Supply ||
		expectedOffering.UnitName != created.UnitName ||
		expectedOffering.BillingType != created.BillingType ||
		expectedOffering.SetupPrice != created.SetupPrice ||
		expectedOffering.UnitPrice != created.UnitPrice ||
		expectedOffering.MinUnits != created.MinUnits ||
		expectedOffering.BillingInterval != created.BillingInterval ||
		expectedOffering.MaxBillingUnitLag != created.MaxBillingUnitLag ||
		expectedOffering.MaxSuspendTime != created.MaxSuspendTime ||
		expectedOffering.MaxInactiveTimeSec != created.MaxInactiveTimeSec ||
		expectedOffering.FreeUnits != created.FreeUnits ||
		!bytes.Equal(expectedOffering.AdditionalParams, created.AdditionalParams) {
		t.Fatal("wrong offering created")
	}
}
