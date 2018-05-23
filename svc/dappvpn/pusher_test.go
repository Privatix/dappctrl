package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/rakyll/statik/fs"

	"github.com/privatix/dappctrl/data"
	_ "github.com/privatix/dappctrl/statik"
	"github.com/privatix/dappctrl/util"
)

const (
	sampleConf        = "/ovpn/samples/server.ovpn"
	sampleCa          = "/ovpn/samples/ca.crt"
	ovpnFileName      = "server.ovpn"
	caFileName        = "ca.crt"
	localConfFileName = "dappvpn.config.json"
	filePerm          = 0644
)

func TestPushConfig(t *testing.T) {
	fxt := data.NewTestFixture(t, testDB)
	defer fxt.Close()

	s := newTestSessSrv(0)
	defer s.Close()

	rootDir, err := ioutil.TempDir("", util.NewUUID())
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(rootDir)

	ovpnConfPath, localConfData := createTestConfigs(t, rootDir,
		fxt.Product.ID, data.TestPassword)

	pushConfig(context.Background(), localConfData, testLogger,
		ovpnConfPath)
}

func readStatFile(t *testing.T, path string) []byte {
	statFS, err := fs.New()
	if err != nil {
		t.Fatal(err)
	}

	f, err := statFS.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	d, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func createTestConfigs(t *testing.T, dir, username,
	password string) (string, *config) {
	stockLocalCfg, err := ioutil.ReadFile(localConfFileName)
	if err != nil {
		t.Fatal(err)
	}

	tempLocalCfg := newConfig()

	if err := json.Unmarshal(stockLocalCfg, tempLocalCfg); err != nil {
		t.Fatal(t)
	}

	cfgData := readStatFile(t, sampleConf)
	caData := readStatFile(t, sampleCa)

	cfgPath := filepath.Join(dir, ovpnFileName)
	caPath := filepath.Join(dir, caFileName)
	localCfgFile := filepath.Join(dir, localConfFileName)

	ioutil.WriteFile(cfgPath, cfgData, filePerm)
	ioutil.WriteFile(caPath, caData, filePerm)

	tempLocalCfg.Pusher.ConfigPath = cfgPath
	tempLocalCfg.Pusher.CaCertPath = caPath
	tempLocalCfg.Server.Username = username
	tempLocalCfg.Server.Password = password
	tempLocalCfg.Server.Addr = testConf.SessionServer.Addr

	if err := writeConfig(localCfgFile, tempLocalCfg); err != nil {
		t.Fatal(err)
	}

	return localCfgFile, tempLocalCfg
}
