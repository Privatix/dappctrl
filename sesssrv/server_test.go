// +build !nosesssrvtest

package sesssrv

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/util/srv"
)

var (
	conf struct {
		DB                *data.DBConfig
		Log               *util.LogConfig
		FileLog           *log.FileConfig
		SessionServer     *Config
		SessionServerTest *testConfig
	}

	db     *reform.DB
	server *Server

	logger2 log.Logger

	allPaths = []string{PathAuth, PathStart, PathStop, PathUpdate,
		PathProductConfig, PathEndpointMsg}
)

type testConfig struct {
	ServerStartupDelay uint // In milliseconds.
	Product            *testProduct
}

type testProduct struct {
	ValidFormatConfig map[string]string
	EmptyConfig       string
}

func newTestConfig() *testConfig {
	return &testConfig{
		ServerStartupDelay: 10,
		Product:            &testProduct{},
	}
}

func newTestFixtures(t *testing.T) *data.TestFixture {
	fixture := data.NewTestFixture(t, db)
	fixture.Channel.ServiceStatus = data.ServiceActive
	if err := db.Update(fixture.Channel); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func TestBadMethod(t *testing.T) {
	client := &http.Client{}
	for _, v := range allPaths {
		req, err := srv.NewHTTPRequest(conf.SessionServer.Config,
			http.MethodPut, v, &srv.Request{Args: nil})
		if err != nil {
			t.Fatalf("failed to create request: %s", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to send request: %s", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Fatalf("unexpected status for bad method: %d",
				resp.StatusCode)
		}
	}
}

func TestBadProductAuth(t *testing.T) {
	fxt := newTestFixtures(t)
	defer fxt.Close()

	for _, v := range allPaths {
		err := Post(conf.SessionServer.Config, logger2,
			"bad-product", "bad-password", v, nil, nil)
		util.TestExpectResult(t, "Post", srv.ErrAccessDenied, err)

		err = Post(conf.SessionServer.Config, logger2,
			fxt.Product.ID, "bad-password", v, nil, nil)
		util.TestExpectResult(t, "Post", srv.ErrAccessDenied, err)
	}
}

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	conf.FileLog = log.NewFileConfig()
	conf.SessionServer = NewConfig()
	conf.SessionServerTest = newTestConfig()
	util.ReadTestConfig(&conf)

	logger := util.NewTestLogger(conf.Log)

	l, err := log.NewStderrLogger(conf.FileLog)
	if err != nil {
		panic(err)
	}

	logger2 = l

	db = data.NewTestDB(conf.DB)
	defer data.CloseDB(db)

	server = NewServer(conf.SessionServer, logger, logger2, db)
	defer server.Close()
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			panic(fmt.Sprintf("failed to serve session "+
				"requests: %s", err))
		}
	}()

	time.Sleep(time.Duration(conf.SessionServerTest.ServerStartupDelay) *
		time.Millisecond)

	os.Exit(m.Run())
}
