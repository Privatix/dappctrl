package config

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/rakyll/statik/fs"
	"io/ioutil"
	"path/filepath"

	_ "github.com/privatix/dappctrl/ept/config/statik"
)

const (
	errGenConfig = "config file is empty"
)

func TestGetText(t *testing.T) {
	out, err := ServerConfig(filepath.Join(samplesPath,
		conf.EptTest.ConfValidCaValid), true,
		conf.EptTest.ExportConfigKeys)
	if err != nil {
		t.Fatal(err)
	}

	param, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}

	address := strings.Split(conf.EptTest.ValidHost[0], ":")

	conf, err := ClientConfig(address[0], address[1], param)
	if err != nil {
		t.Error(err)
	}

	statikFS, err := fs.New()
	if err != nil {
		t.Error(err)
	}

	tpl, err := statikFS.Open(clientTpl)
	if err != nil {
		t.Error(err)
	}
	defer tpl.Close()

	data, err := ioutil.ReadAll(tpl)
	if err != nil {
		t.Error(err)
	}

	result, err := conf.Generate(string(data))
	if err != nil {
		t.Error(err)
	}

	if len(result) == 0 {
		t.Error(errGenConfig)
	}

}
