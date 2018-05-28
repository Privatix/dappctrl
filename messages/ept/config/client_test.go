package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/rakyll/statik/fs"
	"gopkg.in/reform.v1"

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

func checkAccess(t *testing.T, file, login, pass string) {
	d, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(d, []byte(
		fmt.Sprintf("%s\n%s\n", login, pass))) {
		t.Fatal(errDeployConfig)
	}
}

func checkCa(t *testing.T, fileConf string, ca []byte) {
	confData, err := ioutil.ReadFile(fileConf)
	if err != nil {
		t.Fatal(err)
	}

	a := strings.Index(string(confData), `<ca>`)
	b := strings.LastIndex(string(confData), `</ca>`)

	if !reflect.DeepEqual(ca, confData[a+5:b]) {
		t.Fatal(errDeployConfig)
	}
}

func checkConf(t *testing.T, confFile string, srv *srvData) {

	cfg, err := clientConfig(srv.addr, srv.param)
	if err != nil {
		t.Fatal(err)
	}

	cliParams, err := parseConfig(confFile,
		conf.EptTest.ExportConfigKeys, false)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Cipher != cliParams[nameCipher] {
		t.Fatal(errDeployConfig)
	}

	if cfg.ConnectRetry != cliParams[nameConnectRetry] {
		t.Fatal(errDeployConfig)
	}

	if cfg.Ping != cliParams[namePing] {
		t.Fatal(errDeployConfig)
	}

	if cfg.PingRestart != cliParams[namePingRestart] {
		t.Fatal(errDeployConfig)
	}

	if cfg.Proto != cliParams[nameProto] {
		t.Fatal(errDeployConfig)
	}

	if _, ok := cliParams[nameCompLZO]; !ok {
		t.Fatal(errDeployConfig)
	}

	checkCa(t, confFile, []byte(cfg.Ca))
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

	objects := []reform.Record{&data.Channel{ID: util.NewUUID()},
		&data.Endpoint{Channel: util.NewUUID()}}

	for _, record := range objects {
		end, err := Deploy(record, rootDir, srv.addr,
			conf.EptTest.VPNConfig.Login,
			conf.EptTest.VPNConfig.Pass, srv.param)
		if err != nil {
			t.Fatal(err)
		}

		accessFile := filepath.Join(end, clientAccessName)
		confFile := filepath.Join(end, clientConfName)

		if notExist(confFile) ||
			notExist(accessFile) {
			t.Fatal(errDeployConfig)
		}

		checkAccess(t, accessFile, conf.EptTest.VPNConfig.Login,
			conf.EptTest.VPNConfig.Pass)

		checkConf(t, confFile, srv)
	}
}
