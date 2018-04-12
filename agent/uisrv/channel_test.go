// +build !noagentuisrvtest

package uisrv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/privatix/dappctrl/data"

	"github.com/privatix/dappctrl/util"
)

func getChannels(id string) *httptest.ResponseRecorder {
	return getResources(channelsPath,
		map[string]string{"id": id},
		testServer.handleChannels)
}

func testGetChannels(t *testing.T, exp int, id string) {
	res := getChannels(id)
	testGetResources(t, res, exp)
}

func TestGetChannels(t *testing.T) {
	testServer.db.DeleteFrom(data.ChannelTable, "")

	// Get empty list.
	testGetChannels(t, 0, "")

	ch, cleanUp := createTestChannel()
	defer cleanUp()

	// Get all channels.
	testGetChannels(t, 1, "")
	// Get channel by id.
	testGetChannels(t, 1, ch.ID)
	testGetChannels(t, 0, util.NewUUID())
}

func getChannelStatus(id string) *httptest.ResponseRecorder {
	path := fmt.Sprintf("%s%s/status", channelsPath, id)
	r := httptest.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()
	testServer.handleChannels(w, r)
	return w
}

func TestGetChannelStatus(t *testing.T) {
	ch, deleteChan := createTestChannel()
	defer deleteChan()
	// get channel status with a match.
	res := getChannelStatus(ch.ID)
	reply := &statusReply{}
	json.NewDecoder(res.Body).Decode(reply)
	if res.Code != http.StatusOK || reply.Code != http.StatusOK {
		t.Fatalf("failed to get channel status: %d", res.Code)
	}
	if ch.ChannelStatus != reply.Status {
		t.Fatalf("expected %s, got: %s", ch.ChannelStatus, reply.Status)
	}
	// get channel status without a match.
	res = getChannelStatus(util.NewUUID())
	if res.Code != http.StatusNotFound {
		t.Fatalf("expected not found, got: %d", res.Code)
	}
}

func TestUpdateChannelStatus(t *testing.T) {
	// TODO once job queue implemented.
}
