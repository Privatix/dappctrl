package sess_test

import (
	"encoding/json"
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/sess"
	"github.com/privatix/dappctrl/util"
)

func TestGetEndpoint(t *testing.T) {
	fxt := newTestFixture(t)
	defer fxt.Close()

	_, err := handler.GetEndpoint(
		"bad-channel", data.TestPassword, fxt.Channel.ID)
	util.TestExpectResult(t, "GetEndpoint", sess.ErrAccessDenied, err)

	_, err = handler.GetEndpoint(
		fxt.Product.ID, "bad-password", fxt.Channel.ID)
	util.TestExpectResult(t, "GetEndpoint", sess.ErrAccessDenied, err)

	_, err = handler.GetEndpoint(
		fxt.Product.ID, data.TestPassword, "bad-channel")
	util.TestExpectResult(t, "GetEndpoint", sess.ErrChannelNotFound, err)

	ept, err := handler.GetEndpoint(
		fxt.Product.ID, data.TestPassword, fxt.Channel.ID)
	util.TestExpectResult(t, "GetEndpoint", nil, err)

	if ept == nil || ept.ID != fxt.Endpoint.ID {
		t.Error("bad endpoint")
	}
}

func TestSetProductConfig(t *testing.T) {
	fxt := newTestFixture(t)
	defer fxt.Close()

	conf := map[string]string{
		"foo": "foo", "bar": "bar", sess.ProductExternalIP: "1.2.3.4"}

	err := handler.SetProductConfig("bad-channel", data.TestPassword, conf)
	util.TestExpectResult(t, "SetProductConfig", sess.ErrAccessDenied, err)

	err = handler.SetProductConfig(fxt.Product.ID, "bad-password", conf)
	util.TestExpectResult(t, "SetProductConfig", sess.ErrAccessDenied, err)

	err = handler.SetProductConfig(fxt.Product.ID, data.TestPassword, nil)
	util.TestExpectResult(t, "SetProductConfig",
		sess.ErrBadProductConfig, err)

	err = handler.SetProductConfig(fxt.Product.ID, data.TestPassword, conf)
	util.TestExpectResult(t, "SetProductConfig", nil, err)

	data.ReloadFromTestDB(t, db, fxt.Product)

	if fxt.Product.ServiceEndpointAddress == nil ||
		*fxt.Product.ServiceEndpointAddress != "1.2.3.4" {
		t.Error("bad product endpoint address")
	}

	err = json.Unmarshal(fxt.Product.Config, &conf)
	util.TestExpectResult(t, "Unmarshal", nil, err)

	_, addrFound := conf[sess.ProductExternalIP]
	_, fooFound := conf["foo"]
	_, barFound := conf["bar"]
	if len(conf) != 2 || addrFound || !fooFound || !barFound {
		t.Error("bad product config")
	}
}

func TestHeartbeat(t *testing.T) {
	fxt := newTestFixture(t)
	defer fxt.Close()

	_, err := handler.HandleHeartbeat("bad-channel", data.TestPassword)
	util.TestExpectResult(t, "HandleHeartbeat", sess.ErrAccessDenied, err)

	_, err = handler.HandleHeartbeat(fxt.Product.ID, "bad-password")
	util.TestExpectResult(t, "HandleHeartbeat", sess.ErrAccessDenied, err)

	statuses := map[string]string{
		data.ServicePending:     "",
		data.ServiceActivating:  sess.HeartbeatStart,
		data.ServiceActive:      "",
		data.ServiceSuspending:  sess.HeartbeatStop,
		data.ServiceSuspended:   "",
		data.ServiceTerminating: sess.HeartbeatStop,
		data.ServiceTerminated:  "",
	}

	for k, v := range statuses {
		fxt.Channel.ServiceStatus = k
		data.SaveToTestDB(t, db, fxt.Channel)

		result, err := handler.HandleHeartbeat(
			fxt.Product.ID, data.TestPassword)
		util.TestExpectResult(t, "HandleHeartbeat", nil, err)

		expected := sess.HeartbeatResult{Command: v}
		if len(v) != 0 {
			expected.Channel = fxt.Channel.ID
		}

		if result.Command != expected.Command ||
			result.Channel != expected.Channel {
			t.Fatalf("bad result, expected %v, got %v",
				expected, result)
		}
	}
}
