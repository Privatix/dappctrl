package ui_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/AlekSi/pointer"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
)

func TestTopUpChannel(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "TopUpChannel")
	defer fxt.close()

	var j *data.Job
	handler.SetMockQueue(job.QueueMock(func(method int, tx *reform.TX,
		j2 *data.Job, relatedIDs []string, subID string,
		subFunc job.SubFunc) error {
		switch method {
		case job.MockAdd:
			j = j2
		default:
			t.Fatal("unexpected queue call")
		}
		return nil
	}))

	err := handler.TopUpChannel("wrong-password", fxt.Channel.ID, 123)
	assertErrEqual(ui.ErrAccessDenied, err)

	err = handler.TopUpChannel(data.TestPassword, util.NewUUID(), 123)
	assertErrEqual(ui.ErrChannelNotFound, err)

	err = handler.TopUpChannel(data.TestPassword, fxt.Channel.ID, 123)
	assertErrEqual(nil, err)

	if j == nil || j.RelatedType != data.JobChannel ||
		j.RelatedID != fxt.Channel.ID ||
		j.Type != data.JobClientPreChannelTopUp {
		t.Fatalf("expected job not created")
	}

	// Test default gas price setup.
	var testGasPrice uint64 = 500
	deleteSetting := insertDefaultGasPriceSetting(t, testGasPrice)
	defer deleteSetting()
	handler.TopUpChannel(data.TestPassword, fxt.Channel.ID, 0)
	jdata := &data.JobPublishData{}
	json.Unmarshal(j.Data, jdata)
	if jdata.GasPrice != testGasPrice {
		t.Fatal("job with default gas price expected")
	}
}

func TestChangeChannelStatus(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "ChangeChannelStatus")
	defer fxt.close()

	var j *data.Job

	queueMock := job.QueueMock(func(method int, tx *reform.TX,
		j2 *data.Job, relatedIDs []string, subID string,
		subFunc job.SubFunc) error {
		switch method {
		case job.MockAdd:
			j = j2
		default:
			t.Fatal("unexpected queue call")
		}
		return nil
	})

	handler.SetMockQueue(queueMock)
	handler.SetProcessor(proc.NewProcessor(conf.Proc, db, queueMock))

	// Set client channel.
	offering2 := data.NewTestOffering(fxt.User.EthAddr,
		fxt.Product.ID, fxt.TemplateOffer.ID)
	clientChan := data.NewTestChannel(data.NewTestAccount("").EthAddr,
		fxt.Account.EthAddr, offering2.ID, 1, 1,
		data.ChannelActive)
	data.InsertToTestDB(t, db, offering2, clientChan)
	defer data.DeleteFromTestDB(t, db, clientChan, offering2)

	// Test default gas price setup.
	var testGasPrice uint64 = 500
	deleteSetting := insertDefaultGasPriceSetting(t, testGasPrice)
	defer deleteSetting()

	type testObject struct {
		channel       *data.Channel
		action        string
		expJobType    string
		serviceStatus string
	}

	test := func(testData []*testObject) {
		for _, v := range testData {
			v.channel.ServiceStatus = v.serviceStatus
			data.SaveToTestDB(t, db, v.channel)

			err := handler.ChangeChannelStatus(
				data.TestPassword, v.channel.ID, v.action)
			assertErrEqual(nil, err)

			if j == nil || j.Type != v.expJobType ||
				j.RelatedID != v.channel.ID {
				t.Fatal("expected job not created")
			}
		}
	}

	agentTestData := []*testObject{
		{fxt.Channel, ui.ChannelTerminateAction,
			data.JobAgentPreServiceTerminate, data.ServiceActive},
		{fxt.Channel, ui.ChannelPauseAction,
			data.JobAgentPreServiceSuspend, data.ServiceActive},
		{fxt.Channel, ui.ChannelResumeAction,
			data.JobAgentPreServiceUnsuspend,
			data.ServiceSuspended},
	}

	clientTestData := []*testObject{
		{clientChan, ui.ChannelTerminateAction,
			data.JobClientPreServiceTerminate, data.ServiceActive},
		{clientChan, ui.ChannelPauseAction,
			data.JobClientPreServiceSuspend, data.ServiceActive},
		{clientChan, ui.ChannelResumeAction,
			data.JobClientPreServiceUnsuspend,
			data.ServiceSuspended},
		{clientChan, ui.ChannelCloseAction,
			data.JobClientPreUncooperativeCloseRequest,
			data.ServiceActive},
	}

	err := handler.ChangeChannelStatus("wrong-password",
		fxt.Channel.ID, ui.ChannelPauseAction)
	assertErrEqual(ui.ErrAccessDenied, err)

	err = handler.ChangeChannelStatus(data.TestPassword,
		fxt.Channel.ID, "wrong-action")
	assertErrEqual(ui.ErrBadAction, err)

	// Agent side.
	handler.SetMockRole(data.RoleAgent)

	err = handler.ChangeChannelStatus(data.TestPassword, fxt.Channel.ID,
		ui.ChannelCloseAction)
	assertErrEqual(ui.ErrNotAllowedForAgent, err)

	test(agentTestData)

	// Client side.
	handler.SetMockRole(data.RoleClient)

	test(clientTestData)
}

func TestGetAgentChannels(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "GetAgentChannels")
	defer fxt.close()

	assertResult := func(res *ui.GetAgentChannelsResult, err error, exp, total int) {
		assertErrEqual(nil, err)
		if res == nil {
			t.Fatal("result is empty")
		}
		if len(res.Items) != exp {
			t.Fatalf("wanted: %v, got: %v", exp, len(res.Items))
		}
		if res.TotalItems != total {
			t.Fatalf("wanted: %v, got: %v",
				total, res.TotalItems)
		}
	}

	type testObject struct {
		channelStatus string
		serviceStatus string
		expected      int
		offset        uint
		limit         uint
		total         int
	}

	_, err := handler.GetAgentChannels("wrong-password", "", "", 0, 0)
	assertErrEqual(ui.ErrAccessDenied, err)

	testData := []*testObject{
		// Test pagination.
		{"", "", 1, 0, 0, 1},
		{"", "", 1, 0, 1, 1},
		{"", "", 0, 1, 1, 1},
		// Test filtering by channel status and service status.
		{"", "", 1, 0, 0, 1},
		{data.ChannelActive, "", 1, 0, 0, 1},
		{data.ChannelPending, "", 0, 0, 0, 0},
		{"", data.ServicePending, 1, 0, 0, 1},
		{"", data.ServiceActive, 0, 0, 0, 0},
		{data.ChannelActive, data.ServicePending, 1, 0, 0, 1},
		{data.ChannelActive, data.ServiceActive, 0, 0, 0, 0},
	}

	for _, v := range testData {
		res, err := handler.GetAgentChannels(data.TestPassword,
			v.channelStatus, v.serviceStatus, v.offset, v.limit)
		assertResult(res, err, v.expected, v.total)
	}
}

func createClientTestData(t *testing.T, fxt *fixture) (close func()) {
	offering := data.NewTestOffering(fxt.User.EthAddr,
		fxt.Product.ID, fxt.TemplateOffer.ID)
	offering.UnitType = data.UnitScalar
	offering.MaxInactiveTimeSec = pointer.ToUint64(1800)

	channel := data.NewTestChannel(data.NewTestAccount("").EthAddr,
		fxt.Account.EthAddr, offering.ID, 0, 10000, data.ChannelActive)
	channel.ServiceChangedTime = pointer.ToTime(time.Now())

	job2 := data.NewTestJob(data.JobClientPreChannelCreate,
		data.JobUser, data.JobOffering)
	job2.RelatedID = channel.ID
	job2.Status = data.JobDone

	var sessions []*data.Session

	for i := 0; i < int(2); i++ {
		sess := data.NewTestSession(channel.ID)
		sess.LastUsageTime = time.Now()
		sess.SecondsConsumed = 200
		sess.UnitsUsed = 200
		sessions = append(sessions, sess)
	}

	data.InsertToTestDB(t, db, offering, channel, job2, sessions[0],
		sessions[1])

	return func() {
		data.DeleteFromTestDB(t, db, sessions[0], sessions[1],
			job2, channel, offering)
	}
}

func checkGetClientChannelsResult(
	t *testing.T, res *[]ui.ClientChannelInfo) {
	for _, item := range *res {
		var channel data.Channel
		var offering data.Offering
		var job2 data.Job

		if err := db.FindByPrimaryKeyTo(&channel,
			item.ID); err != nil {
			t.Fatal(err)
		}

		if err := db.FindByPrimaryKeyTo(&offering,
			item.Offering); err != nil {
			t.Fatal(err)
		}

		if err := db.FindOneTo(&job2,
			"related_id", item.ID); err != nil {
			t.Fatal(err)
		}

		sess, err := db.FindAllFrom(data.SessionTable,
			"channel", item.ID)
		if err != nil {
			t.Fatal(err)
		}

		var sessions []*data.Session

		for _, v := range sess {
			sessions = append(sessions, v.(*data.Session))
		}

		checkClientChannel(t, &item, channel)
		checkClientChannelStatus(t, &item, channel, offering)
		checkClientChannelJob(t, &item, job2)
		checkClientChannelUsage(t, &item, channel, offering, sessions)
	}
}

func checkClientChannel(
	t *testing.T, resp *ui.ClientChannelInfo, ch data.Channel) {
	if ch.ID != resp.ID {
		t.Fatalf("expected %s, got: %s", ch.ID, resp.ID)
	}

	agent, err := data.HexToAddress(ch.Agent)
	if err != nil {
		t.Fatal(err)
	}

	if agent.String() != resp.Agent {
		t.Fatalf("expected %s, got: %s", agent.String(), resp.Agent)
	}

	client, err := data.HexToAddress(ch.Client)
	if err != nil {
		t.Fatal(err)
	}

	if client.String() != resp.Client {
		t.Fatalf("expected %s, got: %s", client.String(), resp.Client)
	}

	if ch.Offering != resp.Offering {
		t.Fatalf("expected %s, got: %s", ch.Offering, resp.Offering)
	}

	if ch.TotalDeposit != resp.Deposit {
		t.Fatalf("expected %d, got: %d", ch.TotalDeposit, resp.Deposit)
	}
}

func checkClientChannelStatus(t *testing.T, resp *ui.ClientChannelInfo,
	ch data.Channel, offer data.Offering) {
	if ch.ServiceStatus != resp.ChStat.ServiceStatus {
		t.Fatalf("expected %s, got: %s", ch.ServiceStatus,
			resp.ChStat.ServiceStatus)
	}

	if ch.ChannelStatus != resp.ChStat.ChannelStatus {
		t.Fatalf("expected %s, got: %s", ch.ChannelStatus,
			resp.ChStat.ChannelStatus)
	}

	if ch.ServiceChangedTime == nil || resp.ChStat.LastChanged == nil {
		t.Fatal("invalid serviceChangedTime field")
	}

	expectedTime := util.SingleTimeFormat(*ch.ServiceChangedTime)
	if expectedTime != *resp.ChStat.LastChanged {
		t.Fatalf("expected %s, got: %s", expectedTime,
			*resp.ChStat.LastChanged)
	}

	if offer.MaxInactiveTimeSec == nil ||
		*offer.MaxInactiveTimeSec != resp.ChStat.MaxInactiveTime {
		t.Fatal("invalid maxInactiveTime field")
	}
}

func checkClientChannelJob(
	t *testing.T, resp *ui.ClientChannelInfo, job data.Job) {
	if job.ID != resp.Job.ID {
		t.Fatalf("expected %s, got: %s", job.ID, resp.Job.ID)
	}

	if job.Type != resp.Job.Type {
		t.Fatalf("expected %s, got: %s", job.Type, resp.Job.Type)
	}

	if job.Status != resp.Job.Status {
		t.Fatalf("expected %s, got: %s", job.Status, resp.Job.Status)
	}

	if util.SingleTimeFormat(job.CreatedAt) != resp.Job.CreatedAt {
		t.Fatalf("expected %s, got: %s",
			util.SingleTimeFormat(job.CreatedAt), resp.Job.CreatedAt)
	}
}

func checkClientChannelUsage(
	t *testing.T, resp *ui.ClientChannelInfo, ch data.Channel,
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

	if resp.Usage.Current != usage {
		t.Fatalf("expected %d, got: %d",
			usage, resp.Usage.Current)
	}

	cost += usage * offer.UnitPrice
	if resp.Usage.Cost != cost {
		t.Fatalf("expected %d, got: %d",
			cost, resp.Usage.Cost)
	}

	deposit := (ch.TotalDeposit - offer.SetupPrice) /
		offer.UnitPrice
	if resp.Usage.MaxUsage != deposit {
		t.Fatalf("expected %d, got: %d", deposit, resp.Usage.MaxUsage)
	}
}

func TestGetClientChannels(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "GetAgentChannels")
	defer fxt.close()

	// Set client channels.
	cancel := createClientTestData(t, fxt)
	defer cancel()

	cancel2 := createClientTestData(t, fxt)
	defer cancel2()

	assertResult := func(
		res *ui.GetClientChannelsResult, err error, exp, total int) {
		assertErrEqual(nil, err)
		if res == nil {
			t.Fatal("result is empty")
		}

		if len(res.Items) != exp {
			t.Fatalf("wanted: %v, got: %v", exp, len(res.Items))
		}

		if len(res.Items) != 0 {
			checkGetClientChannelsResult(t, &res.Items)
		}

		if res.TotalItems != total {
			t.Fatalf("wanted: %v, got: %v", total, res.TotalItems)
		}
	}

	_, err := handler.GetClientChannels("wrong-password", "", "", 0, 0)
	assertErrEqual(ui.ErrAccessDenied, err)

	type testObject struct {
		channelStatus string
		serviceStatus string
		expected      int
		offset        uint
		limit         uint
		total         int
	}

	testData := []*testObject{
		// Test pagination.
		{"", "", 2, 0, 0, 2},
		{"", "", 1, 0, 1, 2},
		{"", "", 1, 1, 2, 2},
		{"", "", 0, 2, 2, 2},
		// Test filtering by channel status and service status.
		{data.ChannelActive, "", 2, 0, 0, 2},
		{data.ChannelPending, "", 0, 0, 0, 0},
		{"", data.ServicePending, 2, 0, 0, 2},
		{"", data.ServiceActive, 0, 0, 0, 0},
		{data.ChannelActive, data.ServicePending, 2, 0, 0, 2},
		{data.ChannelActive, data.ServiceActive, 0, 0, 0, 0},
	}

	for _, v := range testData {
		res, err := handler.GetClientChannels(data.TestPassword,
			v.channelStatus, v.serviceStatus, v.offset, v.limit)
		assertResult(res, err, v.expected, v.total)
	}
}
