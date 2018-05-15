// +build !noagentuisrvtest

package uisrv

import (
	"net/http"
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func getSessions(t *testing.T, chanID string) *http.Response {
	return getResources(t, sessionsPath, map[string]string{
		"channelId": chanID,
	})
}

func testGetSessions(t *testing.T, exp int, chanID string) {
	res := getSessions(t, chanID)
	testGetResources(t, res, exp)
}

func TestGetSessions(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	// Get empty list.
	testGetSessions(t, 0, "")

	// Get all.
	ch := createTestChannel(t)
	sess := data.NewTestSession(ch.ID)
	insertItems(t, sess)
	testGetSessions(t, 1, "")

	// Get by channel id.
	testGetSessions(t, 1, sess.Channel)
	testGetSessions(t, 0, util.NewUUID())
}
