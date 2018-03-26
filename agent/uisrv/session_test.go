// +build !noagentuisrvtest

package uisrv

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func getSessions(chanID string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", sessionsPath, nil)
	if chanID != "" {
		r.Form = make(url.Values)
		r.Form.Add("channelId", chanID)
	}
	w := httptest.NewRecorder()
	testServer.handleSessions(w, r)
	return w
}

func testGetSessions(t *testing.T, exp int, chanID string) {
	res := getSessions(chanID)
	testGetResources(t, res, exp)
}

func TestGetSessions(t *testing.T) {
	// Get empty list.
	testGetSessions(t, 0, "")
	// Get all.
	ch, deleteChan := createTestChannel()
	sess := data.NewTestSession(ch.ID)
	deleteSess := insertItems(sess)
	defer deleteChan()
	defer deleteSess()
	testGetSessions(t, 1, "")
	// Get by channel id.
	testGetSessions(t, 1, sess.Channel)
	testGetSessions(t, 0, util.NewUUID())
}
