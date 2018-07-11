package msg

import (
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/svc/dappvpn/mon"
	"github.com/privatix/dappctrl/util"
)

var (
	conf struct {
		DB                *data.DBConfig
		Log               *util.LogConfig
		SessionServer     *sesssrv.Config
		SessionServerTest *testSessSrvConfig
		VPNConfigPusher   *pusherTestConf
		VPNMonitor        *mon.Config
	}

	db     *reform.DB
	logger *util.Logger

	parameters map[string]string
)

type pusherTestConf struct {
	ExportConfigKeys []string
	TestConfig       map[string]string
	TimeOut          int64
}

type testSessSrv struct {
	server *sesssrv.Server
}

type testSessSrvConfig struct {
	ServerStartupDelay uint
}

func newPusherTestConf() *pusherTestConf {
	return &pusherTestConf{
		TestConfig: make(map[string]string),
	}
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

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	conf.SessionServer = sesssrv.NewConfig()
	conf.SessionServerTest = newSessSrvTestConfig()
	conf.VPNConfigPusher = newPusherTestConf()
	conf.VPNMonitor = mon.NewConfig()

	util.ReadTestConfig(&conf)

	logger = util.NewTestLogger(conf.Log)

	db = data.NewTestDB(conf.DB, logger)
	defer data.CloseDB(db)

	if len(conf.VPNConfigPusher.TestConfig) == 0 {
		log.Fatal("empty test config")
	}

	parameters = conf.VPNConfigPusher.TestConfig

	os.Exit(m.Run())
}
