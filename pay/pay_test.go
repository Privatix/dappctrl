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
)

const testPassword = "test-password"

type testFixture struct {
	clientAcc *data.Account
	client    *data.User
	agent     *data.Account
	offering  *data.Offering
	channel   *data.Channel
}

func newFixture(t *testing.T) *testFixture {
	clientAcc := data.NewTestAccount(testPassword)

	client := data.NewTestUser()
	client.PublicKey = clientAcc.PublicKey
	client.EthAddr = clientAcc.EthAddr

	agent := data.NewTestAccount(testPassword)

	product := data.NewTestProduct()

	template := data.NewTestTemplate(data.TemplateOffer)

	offering := data.NewTestOffering(agent.EthAddr,
		product.ID, template.ID)

	channel := data.NewTestChannel(agent.EthAddr,
		client.EthAddr, offering.ID, 0, 100,
		data.ChannelActive)

	data.InsertToTestDB(t, testDB, client, agent, product, template,
		offering, channel)

	return &testFixture{clientAcc, client, agent, offering, channel}
}

func newTestPayload(amount uint64, channel *data.Channel,
	offering *data.Offering, clientAcc *data.Account) *payload {
	pld := &payload{
		AgentAddress:    channel.Agent,
		OpenBlockNumber: channel.Block,
		OfferingHash:    offering.Hash,
		Balance:         amount,
		ContractAddress: "<contract address>",
	}
	sig, err := clientAcc.Sign(hash(pld), testPassword)
	if err != nil {
		panic(err)
	}
	pld.BalanceMsgSig = data.FromBytes(sig)
	return pld
}

func sendTestRequest(pld *payload) *httptest.ResponseRecorder {
	body := &bytes.Buffer{}
	json.NewEncoder(body).Encode(pld)
	r := httptest.NewRequest(http.MethodPost, payPath, body)
	w := httptest.NewRecorder()
	util.ValidateMethod(testServer.handlePay, http.MethodPost)(w, r)
	return w
}

func TestValidPayment(t *testing.T) {
	defer data.CleanTestDB(t, testDB)
	fixture := newFixture(t)

	// 100 is a test payment amount
	payload := newTestPayload(100, fixture.channel, fixture.offering, fixture.clientAcc)
	w := sendTestRequest(payload)
	if w.Code != http.StatusOK {
		t.Errorf("expect response ok, got: %d", w.Code)
		t.Log(w.Body)
	}

	updated := &data.Channel{}
	err := testDB.FindOneTo(updated, "block", payload.OpenBlockNumber)
	if err != nil {
		panic(err)
	}

	if updated.ReceiptSignature != payload.BalanceMsgSig {
		t.Error("receipt signature is not updated")
	}

	if updated.ReceiptBalance != payload.Balance {
		t.Error("receipt balance is not updated")
	}
}

func TestInvalidPayments(t *testing.T) {
	defer data.CleanTestDB(t, testDB)
	fixture := newFixture(t)

	validPayload := newTestPayload(1, fixture.channel, fixture.offering, fixture.clientAcc)
	wrongBlock := &payload{
		AgentAddress:    validPayload.AgentAddress,
		OpenBlockNumber: validPayload.OpenBlockNumber + 1,
		OfferingHash:    validPayload.OfferingHash,
		Balance:         validPayload.Balance,
		BalanceMsgSig:   validPayload.BalanceMsgSig,
		ContractAddress: validPayload.ContractAddress,
	}

	closedChannel := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, fixture.offering.ID, 0, 100,
		data.ChannelClosedCoop)

	validCh := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, fixture.offering.ID, 10, 100,
		data.ChannelActive)

	data.InsertToTestDB(t, testDB, closedChannel, validCh)

	closedState := newTestPayload(1, closedChannel, fixture.offering, fixture.clientAcc)

	lessBalance := newTestPayload(9, validCh, fixture.offering, fixture.clientAcc)

	overcharging := newTestPayload(100+1, validCh, fixture.offering, fixture.clientAcc)

	otherUser := data.NewTestAccount(testPassword)
	otherUsersSignature := newTestPayload(100, validCh, fixture.offering, otherUser)

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
		if w.Code == http.StatusOK {
			t.Logf("response: %d, %s", w.Code, w.Body)
			t.Logf("payload: %+v\n", pld)
			t.Errorf("expected server to fail, got: %d", w.Code)
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

	os.Exit(m.Run())
}
