// +build !nopaymenttest

package pay

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

var (
	testServer *Server
	testDB     *reform.DB
	testData   struct {
		client   *data.Account
		agent    *data.User
		offering *data.Offering
		channel  *data.Channel
	}
)

func newTestPayload(amount uint64, ch *data.Channel,
	client *data.Account) *payload {
	pld := &payload{
		AgentAddress:    "<agent address>",
		OpenBlockNumber: ch.Block,
		OfferingHash:    "<offering hash>",
		Balance:         amount,
		ContractAddress: "<contract address>",
	}
	sig, err := client.Sign(hash(pld))
	if err != nil {
		panic(err)
	}
	pld.BalanceMsgSig = data.FromBytes(sig)
	return pld
}

func sendTestRequest(pld *payload) *httptest.ResponseRecorder {
	body := &bytes.Buffer{}
	json.NewEncoder(body).Encode(pld)
	r := httptest.NewRequest("POST", payPath, body)
	w := httptest.NewRecorder()
	util.ValidateMethod(testServer.handlePay, "POST")(w, r)
	return w
}

func TestValidPayment(t *testing.T) {
	// 100 is a test payment amount
	pld := newTestPayload(100, testData.channel, testData.client)
	w := sendTestRequest(pld)
	if w.Code != http.StatusOK {
		t.Errorf("expect response ok, got: %d", w.Code)
		t.Log(w.Body)
	}
	updated := &data.Channel{}
	if err := testDB.FindOneTo(updated, "block", pld.OpenBlockNumber); err != nil {
		panic(err)
	}
	if updated.ReceiptSignature != pld.BalanceMsgSig {
		t.Error("receipt signature is not updated")
	}
	if updated.ReceiptBalance != pld.Balance {
		t.Error("receipt balance is not updated")
	}
}

func TestInvalidPayments(t *testing.T) {
	validPayload := newTestPayload(1, testData.channel, testData.client)
	wrongBlock := &payload{
		AgentAddress:    validPayload.AgentAddress,
		OpenBlockNumber: validPayload.OpenBlockNumber + 1,
		OfferingHash:    validPayload.OfferingHash,
		Balance:         validPayload.Balance,
		BalanceMsgSig:   validPayload.BalanceMsgSig,
		ContractAddress: validPayload.ContractAddress,
	}

	closedChannel := data.NewTestChannel(testData.agent.ID, testData.client.ID,
		testData.offering.ID, 0, 100, data.ChannelClosedCoop)
	testDB.Insert(closedChannel)
	defer func() { testDB.Delete(closedChannel) }()
	closedState := newTestPayload(1,
		closedChannel,
		testData.client)

	validCh := data.NewTestChannel(testData.agent.ID, testData.client.ID,
		testData.offering.ID, 10, 100, data.ChannelActive)
	testDB.Insert(validCh)
	defer func() { testDB.Delete(validCh) }()
	lessBalance := newTestPayload(9, validCh, testData.client)

	overcharging := newTestPayload(100+1, validCh, testData.client)

	otherUsersSignature := newTestPayload(100, validCh, data.NewTestAccount())

	for _, pld := range []*payload{
		// wrong block number
		wrongBlock,
		// channel state is "closed_coop"
		closedState,
		// balance is less then last given
		lessBalance,
		// balance is greater then total_deposit
		overcharging,
		// signature doesn't correspond to channels user
		otherUsersSignature,
	} {
		w := sendTestRequest(pld)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected server to reply with %d, got: %d",
				http.StatusUnauthorized, w.Code)
			t.Logf("response: %s", w.Body)
		}
	}
}

func TestMain(m *testing.M) {
	var conf struct {
		DB  *data.DBConfig
		Log *util.LogConfig
	}
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	util.ReadTestConfig(&conf)
	logger := util.NewTestLogger(conf.Log)
	testDB = data.NewTestDB(conf.DB, logger)
	defer data.CloseDB(testDB)
	testServer = NewServer(nil, logger, testDB)

	// prepare test data
	testData.client = data.NewTestAccount()
	testData.agent = data.NewTestUser()
	prt := data.NewTestProduct()
	tpl := data.NewTestTemplate(data.TemplateOffer)
	testData.offering = data.NewTestOffering(testData.agent.EthAddr, prt.ID, tpl.ID)
	testData.channel = data.NewTestChannel(testData.agent.EthAddr, testData.client.EthAddr,
		testData.offering.ID, 0, 100, data.ChannelActive)

	data.InsertItems(testDB,
		testData.client, &data.User{
			ID:        util.NewUUID(),
			EthAddr:   testData.client.EthAddr,
			PublicKey: testData.client.PublicKey,
		},
		testData.agent,
		prt,
		tpl,
		testData.offering,
		testData.channel)

	exitcode := m.Run()

	// clean up
	data.CleanDB(testDB)

	os.Exit(exitcode)
}
