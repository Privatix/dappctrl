package pusher

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rakyll/statik/fs"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/sesssrv"
	_ "github.com/privatix/dappctrl/statik"
	"github.com/privatix/dappctrl/util"
)

const (
	sampleConf   = "/ovpn/samples/server.ovpn"
	sampleCa     = "/ovpn/samples/ca.crt"
	ovpnFileName = "server.ovpn"
	caFileName   = "ca.crt"
	filePerm     = 0644
)

var (
	conf struct {
		DB                *data.DBConfig
		Log               *util.LogConfig
		SessionServer     *sesssrv.Config
		SessionServerTest *testSessSrvConfig
		VPNConfigPusher   *pusherTestConf
	}

	db     *reform.DB
	logger *util.Logger
)

type pusherTestConf struct {
	ExportConfigKeys []string
	TimeOut          int64
}

type testSessSrv struct {
	server *sesssrv.Server
}

type testSessSrvConfig struct {
	ServerStartupDelay uint
}

func newPusherTestConf() *pusherTestConf {
	return &pusherTestConf{}
}

func newTestSessSrv(t *testing.T, timeout time.Duration) *testSessSrv {
	s := sesssrv.NewServer(conf.SessionServer, logger, db)
	go func() {
		time.Sleep(timeout)
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			t.Fatalf("failed to serve session requests: %s", err)
		}
	}()

	time.Sleep(time.Duration(
		conf.SessionServerTest.ServerStartupDelay) * time.Millisecond)

	return &testSessSrv{s}
}

func (s testSessSrv) Close() {
	s.server.Close()
}

func newSessSrvTestConfig() *testSessSrvConfig {
	return &testSessSrvConfig{
		ServerStartupDelay: 10,
	}
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

func createTestConfig(t *testing.T, dir string) *Config {
	cfgData := readStatFile(t, sampleConf)
	caData := readStatFile(t, sampleCa)

	cfgPath := filepath.Join(dir, ovpnFileName)
	caPath := filepath.Join(dir, caFileName)

	if err := ioutil.WriteFile(cfgPath, cfgData, filePerm); err != nil {
		t.Fatal(err)
	}

	if err := ioutil.WriteFile(caPath, caData, filePerm); err != nil {
		t.Fatal(err)
	}

	return &Config{
		ExportConfigKeys: conf.VPNConfigPusher.ExportConfigKeys,
		ConfigPath:       cfgPath,
		CaCertPath:       caPath,
		Pushed:           false,
		TimeOut:          conf.VPNConfigPusher.TimeOut,
	}
}

func TestPushConfig(t *testing.T) {
	fxt := data.NewTestFixture(t, db)
	defer fxt.Close()

	s := newTestSessSrv(t, 0)
	defer s.Close()

	rootDir, err := ioutil.TempDir("", util.NewUUID())
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(rootDir)

	c := NewCollect(createTestConfig(t, rootDir),
		conf.SessionServer.Config, fxt.Product.ID,
		data.TestPassword, logger)

	if err := PushConfig(context.Background(), c); err != nil {
		t.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	conf.SessionServer = sesssrv.NewConfig()
	conf.SessionServerTest = newSessSrvTestConfig()
	conf.VPNConfigPusher = newPusherTestConf()

	util.ReadTestConfig(&conf)

	logger = util.NewTestLogger(conf.Log)

	db = data.NewTestDB(conf.DB, logger)
	defer data.CloseDB(db)

	os.Exit(m.Run())
}
