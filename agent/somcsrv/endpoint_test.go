package somcsrv_test

import (
	"testing"

	"github.com/privatix/dappctrl/agent/somcsrv"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func TestGetEndpointMessage(t *testing.T) {
	// Test channel does not exist.
	_, err := handler.Endpoint("my-channel-key")
	util.TestExpectResult(t, "Endpoint", somcsrv.ErrChannelNotFound, err)

	// Test endpoint exists.
	fxt := data.NewTestFixture(t, db)
	defer fxt.Close()

	ch := fxt.Channel
	keyBytes, err := (data.ChannelKey(
		ch.Client, ch.Agent, ch.Block, fxt.Offering.Hash))
	util.TestExpectResult(t, "ChannelKey", nil, err)
	channelKey := data.FromBytes(keyBytes)

	// TODO: store channel key in table for fast lookup?
	rawMsg, err := handler.Endpoint(channelKey)
	util.TestExpectResult(t, "Endpoint", nil, err)
	if fxt.Endpoint.RawMsg != *rawMsg {
		t.Fatalf("wanted: %s, go: %s", fxt.Endpoint.RawMsg, *rawMsg)
	}

	// Test endpoint does not exist.
	data.DeleteFromTestDB(t, db, fxt.Endpoint)
	// Restore record for proper fixture close.
	defer data.SaveToTestDB(t, db, fxt.Endpoint)
	_, err = handler.Endpoint(channelKey)
	util.TestExpectResult(t, "Endpoint", somcsrv.ErrEndpointNotFound, err)
}
