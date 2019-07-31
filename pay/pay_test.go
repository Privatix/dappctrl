// +build !nopaytest

package pay

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/util/srv"
)

var (
	testServer *Server
	testDB     *reform.DB
	conf       struct {
		DB            *data.DBConfig
		Log           *log.WriterConfig
		PayServer     *Config
		PayServerTest struct {
			ServerStartupDelay uint // In milliseconds.
		}
		Job  *job.Config
		Proc *proc.Config
	}
	pr *proc.Processor
)

func newFixture(t *testing.T) *data.TestFixture {
	fxt := data.NewTestFixture(t, testDB)
	fxt.Channel.TotalDeposit = 99999
	data.SaveToTestDB(t, testDB, fxt.Channel)
	return fxt
}

func newTestPayload(t *testing.T, amount uint64, channel *data.Channel,
	offering *data.Offering, clientAcc *data.Account) *paymentPayload {

	testPSCAddr := common.HexToAddress("0x1")

	pld := &paymentPayload{
		AgentAddress:    channel.Agent,
		OpenBlockNumber: channel.Block,
		OfferingHash:    offering.Hash,
		Balance:         amount,
		ContractAddress: data.HexFromBytes(testPSCAddr.Bytes()),
	}

	agentAddr := data.TestToAddress(t, channel.Agent)

	offeringHash := data.TestToHash(t, pld.OfferingHash)

	hash := eth.BalanceProofHash(testPSCAddr, agentAddr, pld.OpenBlockNumber,
		offeringHash, pld.Balance)

	key, err := data.TestToPrivateKey(clientAcc.PrivateKey, data.TestPassword)
	util.TestExpectResult(t, "to private key", nil, err)

	sig, err := crypto.Sign(hash, key)
	util.TestExpectResult(t, "sign", nil, err)

	pld.BalanceMsgSig = data.FromBytes(sig)

	return pld
}

func sendTestRequest(t *testing.T, pld *paymentPayload) *srv.Response {
	data, err := json.Marshal(pld)
	if err != nil {
		t.Fatalf("%v, %v", err, util.Caller())
	}
	req, err := srv.NewHTTPRequest(conf.PayServer.Config, http.MethodPost, payPath,
		&srv.Request{Args: data})
	if err != nil {
		t.Fatalf("%v, %v", err, util.Caller())
	}
	response, err := srv.Send(req)
	if err != nil {
		t.Fatalf("%v, %v", err, util.Caller())
	}
	return response
}

func TestValidPayment(t *testing.T) {
	defer data.CleanTestDB(t, testDB)
	fxt := newFixture(t)

	// 100 is a test payment amount
	payload := newTestPayload(t,
		100, fxt.Channel, fxt.Offering, fxt.UserAcc)
	res := sendTestRequest(t, payload)
	if res.Error != nil {
		t.Fatal(res.Error)
	}

	updated := &data.Channel{}
	data.FindInTestDB(
		t, testDB, updated, "block", payload.OpenBlockNumber)

	if *updated.ReceiptSignature != payload.BalanceMsgSig {
		t.Error("receipt signature is not updated")
	}

	if updated.ReceiptBalance != payload.Balance {
		t.Error("receipt balance is not updated")
	}
}

func TestInvalidPayments(t *testing.T) {
	defer data.CleanTestDB(t, testDB)
	fxt := newFixture(t)

	validPayload := newTestPayload(
		t, 1, fxt.Channel, fxt.Offering, fxt.UserAcc)
	wrongBlock := &paymentPayload{
		AgentAddress:    validPayload.AgentAddress,
		OpenBlockNumber: validPayload.OpenBlockNumber + 1,
		OfferingHash:    validPayload.OfferingHash,
		Balance:         validPayload.Balance,
		BalanceMsgSig:   validPayload.BalanceMsgSig,
		ContractAddress: validPayload.ContractAddress,
	}

	closedChannel := data.NewTestChannel(fxt.Account.EthAddr,
		fxt.UserAcc.EthAddr, fxt.Offering.ID, 0, 100,
		data.ChannelClosedCoop)

	validCh := data.NewTestChannel(fxt.Account.EthAddr,
		fxt.User.EthAddr, fxt.Offering.ID, 10, 100,
		data.ChannelActive)

	data.InsertToTestDB(t, testDB, closedChannel, validCh)

	closedState := newTestPayload(
		t, 1, closedChannel, fxt.Offering, fxt.UserAcc)

	lessBalance := newTestPayload(
		t, 9, validCh, fxt.Offering, fxt.UserAcc)

	equalBalance := newTestPayload(
		t, 10, validCh, fxt.Offering, fxt.UserAcc)

	overcharging := newTestPayload(
		t, 100+1, validCh, fxt.Offering, fxt.UserAcc)

	otherUser := data.NewTestAccount(data.TestPassword)
	otherUsersSignature := newTestPayload(
		t, 100, validCh, fxt.Offering, otherUser)

	for _, pld := range []*paymentPayload{
		// wrong block number
		wrongBlock,
		// channel state is "closed_coop"
		closedState,
		// balance is less then last given
		lessBalance,
		// balance is equal last given
		equalBalance,
		// balance is greater then total_deposit
		overcharging,
		// signature doesn't correspond to channels user
		otherUsersSignature,
	} {
		res := sendTestRequest(t, pld)
		if res.Error == nil {
			t.Fatal("exected error, got success")
		}
	}
}

func TestServiceTerminate(t *testing.T) {
	defer data.CleanTestDB(t, testDB)
	fxt := newFixture(t)

	fxt.Channel.ServiceStatus = data.ServiceActive
	path := srv.GetURL(conf.PayServer.Config, payPath)

	fxt.Endpoint.PaymentReceiverAddress = &path

	data.SaveToTestDB(t, testDB, fxt.Channel, fxt.Endpoint)

	mock := func(req *http.Request) (*srv.Response, error) {
		return &srv.Response{
			Error: &srv.Error{
				Code: errCodeTerminatedService,
			},
		}, nil
	}

	payload := newTestPayload(t,
		100, fxt.Channel, fxt.Offering, fxt.UserAcc)
	postPayload(
		testDB, fxt.Channel, payload, false, 0, pr, mock)

	j := &data.Job{}
	data.FindInTestDB(t, testDB, j, "related_id", fxt.Channel.ID)
	if j.Type != data.JobClientPreServiceTerminate ||
		j.RelatedID != fxt.Channel.ID {
		t.Fatal("wrong job")
	}
}

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.Log = log.NewWriterConfig()
	conf.PayServer = NewConfig()
	conf.Proc = proc.NewConfig()
	args := &util.TestArgs{
		Conf: &conf,
	}
	util.ReadTestArgs(args)

	logger, err := log.NewTestLogger(conf.Log, args.Verbose)
	if err != nil {
		panic(err)
	}

	testDB = data.NewTestDB(conf.DB)
	defer data.CloseDB(testDB)

	testServer = NewServer(conf.PayServer, logger, testDB)
	go func() {
		err := testServer.ListenAndServe()
		if err != http.ErrServerClosed {
			panic("failed to serve session requests: " + err.Error())
		}
	}()

	time.Sleep(time.Duration(conf.PayServerTest.ServerStartupDelay) *
		time.Millisecond)

	queue := job.NewQueue(conf.Job, logger, testDB, nil)
	pr = proc.NewProcessor(conf.Proc, testDB, queue)

	os.Exit(m.Run())
}
