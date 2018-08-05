// +build !nosesssrvtest

package sesssrv

import (
	"bytes"
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func TestEndpointMsg(t *testing.T) {
	fxt := newTestFixtures(t)

	if err := db.Save(fxt.Endpoint); err != nil {
		t.Fatal(err)
	}

	defer fxt.Close()

	args := EndpointMsgArgs{ChannelID: fxt.Channel.ID}

	var ept *data.Endpoint

	err := Post(conf.SessionServer.Config, logger2,
		fxt.Product.ID, data.TestPassword, PathEndpointMsg,
		args, &ept)

	if !bytes.Equal(ept.AdditionalParams, fxt.Endpoint.AdditionalParams) ||
		ept.ServiceEndpointAddress != fxt.Endpoint.ServiceEndpointAddress ||
		ept.Username != fxt.Endpoint.Username ||
		ept.Password != fxt.Endpoint.Password {

	}

	util.TestExpectResult(fxt.T, "Post", nil, err)
}
