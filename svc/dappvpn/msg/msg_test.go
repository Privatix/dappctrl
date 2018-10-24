package msg

import (
	"net/http"
	"os"
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/country"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/svc/dappvpn/mon"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

var (
	conf struct {
		DB                *data.DBConfig
		Log               *util.LogConfig
		StderrLog         *log.WriterConfig
		SessionServer     *sesssrv.Config
		SessionServerTest *testSessSrvConfig
		VPNConfigPusher   *pusherTestConf
		VPNMonitor        *mon.Config
	}

	db     *reform.DB
	logger log.Logger

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

func newTestSessSrv(t *testing.T, timeout time.Duration,
	countryConfig *country.Config) *testSessSrv {
	s := sesssrv.NewServer(conf.SessionServer,
		logger, db, countryConfig, job.NewDummyQueueMock())
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

func newTestCountryConfig(field, url string) *country.Config {
	c := country.NewConfig()
	c.Field = field
	c.URLTemplate = url
	return c
}

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	conf.StderrLog = log.NewWriterConfig()
	conf.SessionServer = sesssrv.NewConfig()
	conf.SessionServerTest = newSessSrvTestConfig()
	conf.VPNConfigPusher = newPusherTestConf()
	conf.VPNMonitor = mon.NewConfig()

	util.ReadTestConfig(&conf)

	l, err := log.NewStderrLogger(conf.StderrLog)
	if err != nil {
		panic(err)
	}

	logger = l

	db = data.NewTestDB(conf.DB)
	defer data.CloseDB(db)

	if len(conf.VPNConfigPusher.TestConfig) == 0 {
		panic("empty test config")
	}

	parameters = conf.VPNConfigPusher.TestConfig

	os.Exit(m.Run())
}
