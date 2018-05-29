// +build !noagentuisrvtest

package uisrv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func getChannels(t *testing.T, id string, agent bool) *http.Response {
	if agent {
		return getResources(t, channelsPath,
			map[string]string{"id": id})
	}
	return getResources(t, clientChannelsPath,
		map[string]string{"id": id})
}

func testGetChannels(t *testing.T, exp int, id string, agent bool) {
	res := getChannels(t, id, agent)
	testGetResources(t, res, exp)
}

func TestGetChannels(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	createAcc := func(t *testing.T, ch *data.Channel, agent bool) {
		acc := data.NewTestAccount(testPassword)
		if agent {
			acc.EthAddr = ch.Agent
		} else {
			acc.EthAddr = genEthAddr(t)
		}
		insertItems(t, acc)
	}

	// Get empty list.
	testGetChannels(t, 0, "", true)

	chAgent := createTestChannel(t)
	chClient := createTestChannel(t)

	createAcc(t, chAgent, true)
	createAcc(t, nil, false)

	// Get all channels for Agent and Client.
	testGetChannels(t, 1, "", true)
	testGetChannels(t, 1, "", false)

	// Get channel by id.
	testGetChannels(t, 1, chAgent.ID, true)
	testGetChannels(t, 1, chClient.ID, false)
	testGetChannels(t, 0, util.NewUUID(), true)
	testGetChannels(t, 0, util.NewUUID(), false)
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

func sendChannelAction(t *testing.T, id, action string) *http.Response {
	path := fmt.Sprint(channelsPath, id, "/status")
	payload := &ActionPayload{Action: action}
	return sendPayload(t, http.MethodPut, path, payload)
}

func TestUpdateChannelStatus(t *testing.T) {
	fixture := data.NewTestFixture(t, testServer.db)
	defer fixture.Close()
	defer setTestUserCredentials(t)()

	testJobCreated := func(action string, jobType string) {
		res := sendChannelAction(t, fixture.Channel.ID, action)
		if res.StatusCode != http.StatusOK {
			t.Fatal("got: ", res.Status)
		}
		jobTerm := &data.Job{}
		data.FindInTestDB(t, testServer.db, jobTerm, "type", jobType)
		data.DeleteFromTestDB(t, testServer.db, jobTerm)
	}

	res := sendChannelAction(t, fixture.Channel.ID, "wrong-action")
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("wanted: %d, got: %v", http.StatusBadRequest, res.Status)
	}

	testJobCreated(channelTerminate, data.JobAgentPreServiceTerminate)
	testJobCreated(channelPause, data.JobAgentPreServiceSuspend)
	testJobCreated(channelResume, data.JobAgentPreServiceUnsuspend)
}
