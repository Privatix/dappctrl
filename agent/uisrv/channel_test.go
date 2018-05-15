// +build !noagentuisrvtest

package uisrv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/privatix/dappctrl/util"
)

func getChannels(t *testing.T, id string) *http.Response {
	return getResources(t, channelsPath,
		map[string]string{"id": id})
}

func testGetChannels(t *testing.T, exp int, id string) {
	res := getChannels(t, id)
	testGetResources(t, res, exp)
}

func TestGetChannels(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	// Get empty list.
	testGetChannels(t, 0, "")

	ch := createTestChannel(t)

	// Get all channels.
	testGetChannels(t, 1, "")
	// Get channel by id.
	testGetChannels(t, 1, ch.ID)
	testGetChannels(t, 0, util.NewUUID())
}

func getChannelStatus(t *testing.T, id string) *http.Response {
	url := fmt.Sprintf("http://:%s@%s/%s%s/status", testPassword,
		testServer.conf.Addr, channelsPath, id)
	r, err := http.Get(url)
	if err != nil {
		t.Fatal("failed to get channels: ", err)
	}
	return r
}

func TestGetChannelStatus(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	ch := createTestChannel(t)
	// get channel status with a match.
	res := getChannelStatus(t, ch.ID)
	reply := &statusReply{}
	json.NewDecoder(res.Body).Decode(reply)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("failed to get channel status: %d", res.StatusCode)
	}
	if ch.ChannelStatus != reply.Status {
		t.Fatalf("expected %s, got: %s", ch.ChannelStatus, reply.Status)
	}
	// get channel status without a match.
	res = getChannelStatus(t, util.NewUUID())
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected not found, got: %d", res.StatusCode)
	}
}

func TestUpdateChannelStatus(t *testing.T) {
	// TODO once job queue implemented.
}
