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

	"github.com/AlekSi/pointer"
	"github.com/rakyll/statik/fs"

	"github.com/privatix/dappctrl/data"
	_ "github.com/privatix/dappctrl/statik"
	"github.com/privatix/dappctrl/util"
)

const (
	errGenConfig    = "config file is empty"
	errDeployConfig = "error deploy config"
	errTestAddress  = "test network address are not available"

	nameCipher         = "cipher"
	nameConnectRetry   = "connect-retry"
	nameManagementPort = "management"
	namePing           = "ping"
	namePingRestart    = "ping-restart"
	nameServerAddress  = "serverAddress"

	testManagementPort = 1234
	testServerName     = "testserver"
)

type srvData struct {
	addr       string
	param      []byte
	managePort uint16
}

func createSrvData(t *testing.T) *srvData {
	out := srvConfig(t)
	out[nameServerAddress] = testServerName

	param, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}

	if len(conf.EptTest.ValidHost) < 1 {
		t.Fatal(errTestAddress)
	}

	address := strings.Split(conf.EptTest.ValidHost[0], ":")

	return &srvData{address[0], param, defaultManagementPort}
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

func checkConf(t *testing.T, confFile string, srv *srvData, keys []string) {

	cfg, err := clientConfig(srv.addr, srv.param, srv.managePort)
	if err != nil {
		t.Fatal(err)
	}

	cliParams, err := parseConfig(confFile, keys, false)
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

	if _, ok := cliParams[nameManagementPort]; !ok {
		t.Fatal(errDeployConfig)
	}

	checkCa(t, confFile, []byte(cfg.Ca))
}

func TestGetText(t *testing.T) {
	srv := createSrvData(t)

	conf, err := clientConfig(srv.addr, srv.param, srv.managePort)
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
	fxt := data.NewTestFixture(t, db)
	defer fxt.Close()

	srv := createSrvData(t)

	fxt.Endpoint.AdditionalParams = srv.param
	fxt.Endpoint.ServiceEndpointAddress = pointer.ToString(srv.addr)

	if err := db.Update(fxt.Endpoint); err != nil {
		t.Fatal(err)
	}

	rootDir, err := ioutil.TempDir("", util.NewUUID())
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(rootDir)

	if err := DeployConfig(db, fxt.Endpoint.ID, rootDir,
		testManagementPort); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(rootDir, fxt.Endpoint.Channel)

	accessFile := filepath.Join(target, clientAccessName)
	confFile := filepath.Join(target, clientConfName)

	for _, f := range []string{accessFile, confFile} {
		if notExist(f) {
			t.Fatal(errDeployConfig)
		}
	}

	checkAccess(t, accessFile, *fxt.Endpoint.Username,
		*fxt.Endpoint.Password)

	keys := append(conf.EptTest.ExportConfigKeys, nameManagementPort)

	checkConf(t, confFile, srv, keys)
}
