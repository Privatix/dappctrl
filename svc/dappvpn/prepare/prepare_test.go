package prepare

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/svc/dappvpn/config"
	"github.com/privatix/dappctrl/svc/dappvpn/mon"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

const (
	accessFileName = "access.ovpn"
	configFileName = "client.ovpn"
)

var (
	conf struct {
		DB                *data.DBConfig
		Log               *util.LogConfig
		StderrLog         *log.WriterConfig
		SessionServer     *sesssrv.Config
		SessionServerTest *testSessSrvConfig
		VPNMonitor        *mon.Config
	}

	db     *reform.DB
	logger log.Logger
)

type testSessSrvConfig struct {
	ServerStartupDelay uint
}

type testSessSrv struct {
	server *sesssrv.Server
}

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	conf.StderrLog = log.NewWriterConfig()
	conf.SessionServer = sesssrv.NewConfig()
	conf.SessionServerTest = newSessSrvTestConfig()

	util.ReadTestConfig(&conf)

	l, err := log.NewStderrLogger(conf.StderrLog)
	if err != nil {
		panic(err)
	}

	logger = l

	db = data.NewTestDB(conf.DB)
	defer data.CloseDB(db)

	os.Exit(m.Run())
}

func newTestSessSrv(t *testing.T, timeout time.Duration) *testSessSrv {
	s := sesssrv.NewServer(conf.SessionServer, logger, db, nil)
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

func configDestination(dir string) string {
	return filepath.Join(dir, configFileName)
}

func accessDestination(dir string) string {
	return filepath.Join(dir, accessFileName)
}

func checkFile(t *testing.T, file string) {
	stat, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}

	if stat.Size() == 0 {
		t.Fatal("file is empty")
	}
}

func TestClientConfig(t *testing.T) {
	fxt := data.NewTestFixture(t, db)
	defer fxt.Close()

	s := newTestSessSrv(t, 0)
	defer s.Close()

	rootDir, err := ioutil.TempDir("", util.NewUUID())
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(rootDir)

	adapterConfig := config.NewConfig()
	adapterConfig.Server.Config = conf.SessionServer.Config
	adapterConfig.Server.Username = fxt.Product.ID
	adapterConfig.Server.Password = data.TestPassword
	adapterConfig.OpenVPN.ConfigRoot = rootDir
	adapterConfig.Monitor = conf.VPNMonitor
	adapterConfig.FileLog.WriterConfig = conf.StderrLog

	if err := ClientConfig(logger, fxt.Channel.ID,
		adapterConfig); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(rootDir, fxt.Endpoint.Channel)

	checkFile(t, configDestination(target))
	checkFile(t, accessDestination(target))
}
