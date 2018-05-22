package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rakyll/statik/fs"

	"github.com/privatix/dappctrl/data"
	_ "github.com/privatix/dappctrl/statik"
	"github.com/privatix/dappctrl/util"
)

const (
	errGenConfig    = "config file is empty"
	errDeployConfig = "error deploy config"
)

type srvData struct {
	addr  string
	param []byte
}

func createSrvData(t *testing.T) *srvData {
	out := srvConfig(t)

	param, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}

	address := strings.Split(conf.EptTest.ValidHost[0], ":")

	return &srvData{address[0], param}
}

func TestGetText(t *testing.T) {
	srv := createSrvData(t)

	conf, err := clientConfig(srv.addr, srv.param)
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

	d, err := ioutil.ReadAll(tpl)
	if err != nil {
		t.Error(err)
	}

	result, err := conf.generate(string(d))
	if err != nil {
		t.Error(err)
	}

	if len(result) == 0 {
		t.Error(errGenConfig)
	}
}

func TestDeployClientConfig(t *testing.T) {
	srv := createSrvData(t)

	rootDir, err := ioutil.TempDir("", util.NewUUID())
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(rootDir)

	d := NewConfDeployer(rootDir)
	end, err := d.Deploy(&data.Channel{ID: util.NewUUID()},
		srv.addr, conf.EptTest.ConfigTest.Login,
		conf.EptTest.ConfigTest.Pass, srv.param)
	if err != nil {
		t.Fatal(err)
	}

	if isNotExist(filepath.Join(end, clientConfName)) ||
		isNotExist(filepath.Join(end, clientAccessName)) {
		t.Fatal(errDeployConfig)
	}
}
