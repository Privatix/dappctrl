// +build !noagentuisrvtest

package uisrv

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

const (
	errFieldVal = "The field in the response" +
		" does not match the value in the database"
	status = "/status"
)

func getChannels(t *testing.T, params map[string]string,
	agent bool) *http.Response {
	if agent {
		return getResources(t, channelsPath, params)
	}
	return getResources(t, clientChannelsPath, params)
}

func testGetChannels(t *testing.T, exp int,
	params map[string]string, agent bool) {
	res := getChannels(t, params, agent)
	testGetResources(t, res, exp)
}

func getCliChannels(t *testing.T, params map[string]string,
	agent bool) (chs []respGetClientChan) {
	resp := getChannels(t, params, agent)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(bodyBytes,
		&chs); err != nil {
		t.Fatal(err)
	}
	return
}

func testParams(id, channelStatus,
	serviceStatus string) map[string]string {
	return map[string]string{
		"id": id, "channelStatus": channelStatus,
		"serviceStatus": serviceStatus}
}

func testAcc(t *testing.T, ch *data.Channel, agent bool) {
	acc := data.NewTestAccount(testPassword)
	if agent {
		acc.EthAddr = ch.Agent
	} else {
		acc.EthAddr = genEthAddr(t)
	}
	insertItems(t, acc)
}

func testJob(t *testing.T, ch *data.Channel) {
	job := data.NewTestJob(data.JobClientPreChannelCreate,
		data.JobUser, data.JobOffering)
	job.RelatedID = ch.ID
	job.Status = data.JobDone
	insertItems(t, job)
}

func testSess(t *testing.T, chID string, usage, quantity uint64) {
	for i := 0; i < int(quantity); i++ {
		sess := data.NewTestSession(chID)
		sess.LastUsageTime = time.Now()
		sess.SecondsConsumed = usage
		sess.UnitsUsed = usage
		insertItems(t, sess)
	}
}

func checkCliChan(t *testing.T, resp respGetClientChan, ch data.Channel) {
	if ch.ID != resp.ID {
		t.Fatalf("expected %s, got: %s",
			ch.ID, resp.ID)
	}

	if ethAddrFromHex(ch.Agent) != resp.Agent {
		t.Fatalf("expected %s, got: %s",
			ethAddrFromHex(ch.Agent), resp.Agent)
	}

	if ethAddrFromHex(ch.Client) != resp.Client {
		t.Fatalf("expected %s, got: %s",
			ethAddrFromHex(ch.Client), resp.Client)
	}

	if ch.Offering != resp.Offering {
		t.Fatalf("expected %s, got: %s",
			ch.Offering, resp.Offering)
	}

	if ch.TotalDeposit != resp.Deposit {
		t.Fatalf("expected %d, got: %d",
			ch.TotalDeposit, resp.Deposit)
	}
}

func checkCliChanStatus(t *testing.T, resp chanStatusBlock,
	ch data.Channel, offer data.Offering) {
	if ch.ServiceStatus != resp.ServiceStatus {
		t.Fatalf("expected %s, got: %s",
			ch.ServiceStatus, resp.ServiceStatus)
	}

	if ch.ChannelStatus != resp.ChannelStatus {
		t.Fatalf("expected %s, got: %s",
			ch.ChannelStatus, resp.ChannelStatus)
	}

	if ch.ServiceChangedTime == nil || resp.LastChanged == nil {
		t.Fatal(errFieldVal)
	}

	expectedTime := singleTimeFormat(*ch.ServiceChangedTime)
	if expectedTime != *resp.LastChanged {
		t.Fatalf("expected %s, got: %s",
			expectedTime, *resp.LastChanged)
	}

	if offer.MaxInactiveTimeSec == nil ||
		*offer.MaxInactiveTimeSec != resp.MaxInactiveTime {
		t.Fatal(errFieldVal)
	}
}

func checkCliChanJob(t *testing.T, resp jobBlock, job data.Job) {
	if job.ID != resp.ID {
		t.Fatalf("expected %s, got: %s",
			job.ID, resp.ID)
	}

	if job.Type != resp.Type {
		t.Fatalf("expected %s, got: %s",
			job.Type, resp.Type)
	}

	if job.Status != resp.Status {
		t.Fatalf("expected %s, got: %s",
			job.Status, resp.Status)
	}

	if singleTimeFormat(job.CreatedAt) != resp.CreatedAt {
		t.Fatalf("expected %s, got: %s",
			singleTimeFormat(job.CreatedAt),
			resp.CreatedAt)
	}
}

func checkCliChanUsage(t *testing.T, resp usageBlock, ch data.Channel,
	offer data.Offering, sess []*data.Session) {
	var usage uint64
	var cost = offer.SetupPrice

	switch offer.UnitType {
	case data.UnitScalar:
		for _, ses := range sess {
			usage += ses.UnitsUsed
		}
	case data.UnitSeconds:
		for _, ses := range sess {
			usage += ses.SecondsConsumed
		}
	default:
		t.Fatal("unsupported unit type")
	}

	if resp.Current != usage {
		t.Fatalf("expected %d, got: %d",
			usage, resp.Current)
	}

	cost += usage * offer.UnitPrice
	if resp.Cost != cost {
		t.Fatalf("expected %d, got: %d",
			cost, resp.Cost)
	}

	deposit := (ch.TotalDeposit - offer.SetupPrice) /
		offer.UnitPrice
	if resp.MaxUsage != deposit {
		t.Fatalf("expected %d, got: %d", deposit, resp.MaxUsage)
	}
}

func checkRespGetCliChan(t *testing.T, chID string) {
	chs := getCliChannels(t, nil, false)

	for _, cc := range chs {
		if cc.ID != chID {
			continue
		}

		var ch data.Channel
		var offer data.Offering
		var job data.Job

		if err := testServer.db.FindByPrimaryKeyTo(&ch,
			cc.ID); err != nil {
			t.Fatal(err)
		}

		if err := testServer.db.FindByPrimaryKeyTo(&offer,
			cc.Offering); err != nil {
			t.Fatal(err)
		}

		if err := testServer.db.FindOneTo(&job,
			"related_id", ch.ID); err != nil {
			t.Fatal(err)
		}

		sessSlice, err := testServer.db.FindAllFrom(data.SessionTable,
			"channel", ch.ID)
		if err != nil {
			t.Fatal(err)
		}

		var sess []*data.Session

		for _, v := range sessSlice {
			sess = append(sess, v.(*data.Session))
		}

		checkCliChan(t, cc, ch)
		checkCliChanStatus(t, cc.ChStat, ch, offer)
		checkCliChanJob(t, cc.Job, job)
		checkCliChanUsage(t, cc.Usage, ch, offer, sess)
	}

}

func TestGetChannels(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	// Get empty list.
	testGetChannels(t, 0, testParams("", "", ""), true)

	chAgent := createTestChannel(t)
	chClient := createTestChannel(t)

	testAcc(t, chAgent, true)
	testAcc(t, nil, false)
	testJob(t, chClient)
	testSess(t, chClient.ID, 100, 2)
	testSess(t, chAgent.ID, 200, 2)

	// Get all channels for Agent and Client.
	testGetChannels(t, 1, testParams("", "", ""), true)
	testGetChannels(t, 1, testParams("", "", ""), false)

	// Get channel by id.
	testGetChannels(t, 1, testParams(chAgent.ID, "", ""), true)
	testGetChannels(t, 0, testParams(util.NewUUID(), "", ""), true)
	testGetChannels(t, 1, testParams(chClient.ID, "", ""), false)
	testGetChannels(t, 0, testParams(util.NewUUID(), "", ""), false)

	// Get channel by channel status
	testGetChannels(t, 1, testParams("", data.ChannelActive, ""), true)
	testGetChannels(t, 0, testParams("", data.ChannelPending, ""), true)
	testGetChannels(t, 1, testParams("", data.ChannelActive, ""), false)
	testGetChannels(t, 0, testParams("", data.ChannelPending, ""), false)

	// Get channel by service status
	testGetChannels(t, 1, testParams("", "", data.ServicePending), true)
	testGetChannels(t, 0, testParams("", "", data.ServiceActive), true)
	testGetChannels(t, 1, testParams("", "", data.ServicePending), false)
	testGetChannels(t, 0, testParams("", "", data.ServiceActive), false)

	// check response body
	checkRespGetCliChan(t, chClient.ID)
}

func getChannelStatus(t *testing.T, id string, agent bool) *http.Response {
	var path string
	if agent {
		path = channelsPath
	} else {
		path = clientChannelsPath
	}

	url := fmt.Sprintf("http://:%s@%s/%s%s/status", testPassword,
		testServer.conf.Addr, path, id)
	r, err := http.Get(url)
	if err != nil {
		t.Fatal("failed to get channels: ", err)
	}
	return r
}

func checkStatusCode(t *testing.T, resp *http.Response,
	code int, errFormat string) {
	if resp.StatusCode != code {
		t.Fatalf(errFormat, resp.StatusCode)
	}
}

func TestGetAgentChannelStatus(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	ch := createTestChannel(t)
	testAcc(t, ch, true)

	// get channel status with a match.
	res := getChannelStatus(t, ch.ID, true)
	reply := &statusReply{}
	json.NewDecoder(res.Body).Decode(reply)
	checkStatusCode(t, res, http.StatusOK,
		"failed to get channel status: %d")

	if ch.ChannelStatus != reply.Status {
		t.Fatalf("expected %s, got: %s",
			ch.ChannelStatus, reply.Status)
	}

	// get channel status without a match.
	res = getChannelStatus(t, util.NewUUID(), true)
	checkStatusCode(t, res, http.StatusNotFound,
		"expected not found, got: %d")
}

func TestGetClientChannelStatus(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	ch := createTestChannel(t)
	testAcc(t, nil, false)

	var offer data.Offering
	if err := testServer.db.FindByPrimaryKeyTo(&offer,
		ch.Offering); err != nil {
		t.Fatal(err)
	}

	resp := getChannelStatus(t, ch.ID, false)
	result := new(chanStatusBlock)
	json.NewDecoder(resp.Body).Decode(result)
	checkStatusCode(t, resp, http.StatusOK,
		"failed to get channel status: %d")

	checkCliChanStatus(t, *result, *ch, offer)

	// get channel status without a match.
	resp = getChannelStatus(t, util.NewUUID(), false)
	checkStatusCode(t, resp, http.StatusNotFound,
		"expected not found, got: %d")
}

func sendChannelAction(t *testing.T, id, action string, agent bool) *http.Response {
	var path string
	if agent {
		path = channelsPath + id + status
	} else {
		path = clientChannelsPath + id + status
	}

	payload := &ActionPayload{Action: action}
	return sendPayload(t, http.MethodPut, path, payload)
}

func TestUpdateChannelStatus(t *testing.T) {
	fixture := data.NewTestFixture(t, testServer.db)
	defer fixture.Close()
	defer setTestUserCredentials(t)()

	offering2 := data.NewTestOffering(fixture.User.EthAddr,
		fixture.Product.ID, fixture.TemplateOffer.ID)
	clientChan := data.NewTestChannel(data.NewTestAccount("").EthAddr,
		fixture.Account.EthAddr, offering2.ID, 1, 1, data.ChannelActive)
	data.InsertToTestDB(t, testServer.db, offering2, clientChan)
	defer data.DeleteFromTestDB(t, testServer.db, clientChan, offering2)

	testJobCreated := func(action, chanID, jobType string, agent bool) {
		res := sendChannelAction(t, chanID, action, agent)
		job := &data.Job{}
		data.FindInTestDB(t, testServer.db, job, "type", jobType)
		data.DeleteFromTestDB(t, testServer.db, job)
		if res.StatusCode != http.StatusOK {
			s, _ := ioutil.ReadAll(res.Body)
			t.Logf("%s", s)
			t.Fatal("got: ", res.Status)
		}
	}

	testBadRequest := func(chanID string, agent bool) {
		res := sendChannelAction(t, chanID,
			"wrong-action", agent)
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("wanted: %d, got: %v",
				http.StatusBadRequest, res.Status)
		}
	}

	testBadRequest(fixture.Channel.ID, true)
	testBadRequest(clientChan.ID, false)

	testJobCreated(channelTerminate, fixture.Channel.ID,
		data.JobAgentPreServiceTerminate, true)

	fixture.Channel.ServiceStatus = data.ServiceActive
	data.SaveToTestDB(t, testServer.db, fixture.Channel)
	testJobCreated(channelPause, fixture.Channel.ID,
		data.JobAgentPreServiceSuspend, true)

	fixture.Channel.ServiceStatus = data.ServiceSuspended
	data.SaveToTestDB(t, testServer.db, fixture.Channel)
	testJobCreated(channelResume, fixture.Channel.ID,
		data.JobAgentPreServiceUnsuspend, true)

	testJobCreated(channelTerminate, clientChan.ID,
		data.JobClientPreServiceTerminate, false)

	clientChan.ServiceStatus = data.ServiceActive
	data.SaveToTestDB(t, testServer.db, clientChan)
	testJobCreated(channelPause, clientChan.ID,
		data.JobClientPreServiceSuspend, false)

	clientChan.ServiceStatus = data.ServiceSuspended
	data.SaveToTestDB(t, testServer.db, clientChan)
	testJobCreated(channelResume, clientChan.ID,
		data.JobClientPreServiceUnsuspend, false)

	gasPriceSettings := &data.Setting{
		Key:   data.DefaultGasPriceKey,
		Value: "20000000",
	}
	data.InsertToTestDB(t, testServer.db, gasPriceSettings)
	defer data.DeleteFromTestDB(t, testServer.db, gasPriceSettings)

	testJobCreated(clientChannelClose, clientChan.ID,
		data.JobClientPreUncooperativeCloseRequest, false)
}
