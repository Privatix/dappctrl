package ui_test

import (
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
)

type getEndpointsTextData struct {
	channel  string
	template string
	expected int
}

func TestGetEndpoints(t *testing.T) {
	fxt, assertErrEquals := newTest(t, "GetEndpoints")
	defer fxt.close()

	channel := data.NewTestChannel(util.NewUUID(), fxt.Account.ID,
		fxt.Offering.ID, 0, 0, data.ChannelActive)
	endpoint := data.NewTestEndpoint(channel.ID, fxt.TemplateOffer.ID)

	data.InsertToTestDB(t, db, channel, endpoint)
	defer data.DeleteFromTestDB(t, db, endpoint, channel)

	_, err := handler.GetEndpoints("wrong-password", "", "")
	assertErrEquals(ui.ErrAccessDenied, err)

	assertResult := func(res []data.Endpoint, err error, exp int) {
		assertErrEquals(nil, err)
		if len(res) != exp {
			t.Fatalf("wanted: %v, got: %v", exp, len(res))
		}
	}

	testData := []*getEndpointsTextData{
		{"", "", 2},
		{util.NewUUID(), "", 0},
		{fxt.Channel.ID, "", 1},
		{"", util.NewUUID(), 0},
		{"", fxt.TemplateAccess.ID, 1},
		{util.NewUUID(), util.NewUUID(), 0},
		{fxt.Channel.ID, util.NewUUID(), 0},
		{util.NewUUID(), fxt.TemplateAccess.ID, 0},
		{fxt.Channel.ID, fxt.TemplateAccess.ID, 1},
	}

	for _, v := range testData {
		res, err := handler.GetEndpoints(
			data.TestPassword, v.channel, v.template)
		assertResult(res, err, v.expected)
	}
}
