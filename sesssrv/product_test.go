// +build !nosesssrvtest

package sesssrv

import (
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func TestNormalProductConfig(t *testing.T) {
	fxt := newTestFixtures(t)
	defer fxt.Close()

	args := ProductArgs{
		Config: conf.SessionServerTest.Product.ValidFormatConfig}

	err := Post(conf.SessionServer.Config, logger,
		fxt.Product.ID, data.TestPassword, PathProductConfig,
		args, nil)

	util.TestExpectResult(fxt.T, "Post", nil, err)
}

func TestBadProductConfig(t *testing.T) {
	fxt := newTestFixtures(t)
	defer fxt.Close()

	args := ProductArgs{}

	err := Post(conf.SessionServer.Config, logger,
		fxt.Product.ID, data.TestPassword, PathProductConfig,
		args, nil)
	util.TestExpectResult(t, "Post", ErrInvalidProductConf, err)
}

func TestHeartbeat(t *testing.T) {
	fxt := newTestFixtures(t)
	defer fxt.Close()

	statuses := map[string]string{
		data.ServicePending:     "",
		data.ServiceActivating:  HeartbeatStart,
		data.ServiceActive:      "",
		data.ServiceSuspending:  HeartbeatStop,
		data.ServiceSuspended:   "",
		data.ServiceTerminating: HeartbeatStop,
		data.ServiceTerminated:  "",
	}

	for k, v := range statuses {
		fxt.Channel.ServiceStatus = k
		data.SaveToTestDB(t, db, fxt.Channel)

		var result HeartbeatResult
		err := Post(conf.SessionServer.Config,
			logger, fxt.Product.ID, data.TestPassword,
			PathProductHeartbeat, nil, &result)
		util.TestExpectResult(t, "Post", nil, err)

		expected := HeartbeatResult{Command: v}
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
