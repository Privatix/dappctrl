package main

import (
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/util"
)

var (
	testConf struct {
		DB                *data.DBConfig
		Log               *util.LogConfig
		SessionServer     *sesssrv.Config
		SessionServerTest *testConfig
	}

	testDB     *reform.DB
	testLogger *util.Logger
)

type testSessSrv struct {
	server *sesssrv.Server
}

type testConfig struct {
	ServerStartupDelay uint // In milliseconds.
	Product            *testProduct
}

type testProduct struct {
	EmptyConfig       string
	ValidFormatConfig map[string]string
}

func newTestSessSrv(timeout time.Duration) *testSessSrv {
	srv := sesssrv.NewServer(testConf.SessionServer, testLogger, testDB)
	go func() {
		time.Sleep(timeout)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("failed to serve session requests: %s", err)
		}
	}()

	time.Sleep(time.Duration(
		testConf.SessionServerTest.ServerStartupDelay) * time.Millisecond)

	return &testSessSrv{srv}
}

func (s testSessSrv) Close() {
	s.server.Close()
}

func newTestConfig() *testConfig {
	return &testConfig{
		ServerStartupDelay: 10,
		Product:            &testProduct{},
	}
}

func TestMain(m *testing.M) {
	testConf.DB = data.NewDBConfig()
	testConf.Log = util.NewLogConfig()
	testConf.SessionServer = sesssrv.NewConfig()
	testConf.SessionServerTest = newTestConfig()

	util.ReadTestConfig(&testConf)

	testLogger = util.NewTestLogger(testConf.Log)

	testDB = data.NewTestDB(testConf.DB, testLogger)
	defer data.CloseDB(testDB)

	os.Exit(m.Run())
}
