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

	err := Post(conf.SessionServer.Config,
		fxt.Product.ID, data.TestPassword, PathProductConfig,
		args, nil)

	util.TestExpectResult(fxt.T, "Post", nil, err)
}

func TestBadProductConfig(t *testing.T) {
	fxt := newTestFixtures(t)
	defer fxt.Close()

	args := ProductArgs{}

	err := Post(conf.SessionServer.Config,
		fxt.Product.ID, data.TestPassword, PathProductConfig,
		args, nil)
	util.TestExpectResult(t, "Post", ErrInvalidProductConf, err)
}
