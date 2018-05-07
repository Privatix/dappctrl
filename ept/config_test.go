package ept

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/privatix/dappctrl/ept/templates/ovpn"
)

const (
	errGenConfig = "config file is empty"
)

func TestGetText(t *testing.T) {
	out, err := testEMT.ParseConfig(samplesPath +
		conf.EptTest.ConfValidCaValid)
	if err != nil {
		t.Fatal(err.Error())
	}

	param, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}

	address := strings.Split(conf.EptTest.ValidHost[0], ":")

	conf, err := New(address[0], address[1], param)
	if err != nil {
		t.Error(err)
	}

	result, err := conf.GetText(ovpn.ClientConfig)
	if err != nil {
		t.Error(err)
	}

	if len(result) == 0 {
		t.Error(errGenConfig)
	}
}
