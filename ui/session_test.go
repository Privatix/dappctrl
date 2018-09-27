package ui_test

import (
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
)

func TestGetSessions(t *testing.T) {
	fxt, assertErrEquals := newTest(t, "GetSessions")
	defer fxt.close()

	channel := data.NewTestChannel(util.NewUUID(), fxt.Account.ID,
		fxt.Offering.ID, 0, 0, data.ChannelActive)
	session1 := data.NewTestSession(channel.ID)
	session2 := data.NewTestSession(fxt.Channel.ID)

	data.InsertToTestDB(t, db, channel, session1, session2)
	defer data.DeleteFromTestDB(t, db, session2, session1, channel)

	_, err := handler.GetSessions("wrong-password", "")
	assertErrEquals(ui.ErrAccessDenied, err)

	assertResult := func(res []data.Session, err error, exp int) {
		assertErrEquals(nil, err)
		if len(res) != exp {
			t.Fatalf("expected: %d sessions, got: %d",
				exp, len(res))
		}
	}

	res, err := handler.GetSessions(data.TestPassword, "")
	assertResult(res, err, 2)

	res, err = handler.GetSessions(data.TestPassword, fxt.Channel.ID)
	assertResult(res, err, 1)

	res, err = handler.GetSessions(data.TestPassword, util.NewUUID())
	assertResult(res, err, 0)
}
