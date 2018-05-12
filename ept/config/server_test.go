package config

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/util"
)

const (
	errPars     = "incorrect parsing test"
	samplesPath = "samples"
)

var (
	conf struct {
		EptTest           *eptTestConfig
		DB                *data.DBConfig
		Log               *util.LogConfig
		SessionServer     *sesssrv.Config
		SessionServerTest *testConfig
	}

	db     *reform.DB
	logger *util.Logger
)

type testSessSrv struct {
	server *sesssrv.Server
}

type testConfig struct {
	ServerStartupDelay uint // In milliseconds.
	Product            *testProduct
}

type testProduct struct {
	ValidFormatConfig map[string]string
	EmptyConfig       string
}

type eptTestConfig struct {
	Template            string
	ExportConfigKeys    []string
	ValidHash           []string
	InvalidHash         []string
	ValidHost           []string
	InvalidHost         []string
	ConfValidCaValid    string
	ConfInvalid         string
	ConfValidCaInvalid  string
	ConfValidCaEmpty    string
	ConfValidCaNotExist string
	Ca                  string
}

func newTestSessSrv(timeout time.Duration) *testSessSrv {
	srv := sesssrv.NewServer(conf.SessionServer, logger, db)
	go func() {
		time.Sleep(timeout)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("failed to serve session requests: %s", err)
		}
	}()

	time.Sleep(time.Duration(conf.SessionServerTest.ServerStartupDelay) *
		time.Millisecond)

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

func newEptTestConfig() *eptTestConfig {
	return &eptTestConfig{}
}

func validParams(in []string, out map[string]string) bool {
	for _, key := range in {
		delete(out, key)
	}

	delete(out, caPathName)

	if out[caData] == "" {
		return false
	}

	delete(out, caData)

	if len(out) != 0 {
		return false
	}
	return true
}

func joinFile(path, file string) string {
	return filepath.Join(path, file)
}

func TestParsingValidConfig(t *testing.T) {
	out, err := ServerConfig(joinFile(samplesPath,
		conf.EptTest.ConfValidCaValid),
		true, conf.EptTest.ExportConfigKeys)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !validParams(conf.EptTest.ExportConfigKeys, out) {
		t.Fatal(errPars)
	}
}

func TestParsingInvalidConfig(t *testing.T) {
	_, err := ServerConfig(joinFile(samplesPath,
		conf.EptTest.ConfInvalid),
		true, conf.EptTest.ExportConfigKeys)
	if err == nil {
		t.Fatal(errPars)
	}
}

func TestCannotReadCertificateFile(t *testing.T) {
	_, err := ServerConfig(joinFile(samplesPath,
		conf.EptTest.ConfValidCaNotExist),
		true, conf.EptTest.ExportConfigKeys)
	if err == nil {
		t.Fatal(errPars)
	}
}

func TestCertificateIsEmpty(t *testing.T) {
	_, err := ServerConfig(joinFile(samplesPath,
		conf.EptTest.ConfValidCaEmpty),
		true, conf.EptTest.ExportConfigKeys)
	if err == nil {
		t.Fatal(errPars)
	}
}

func TestInvalidCertificate(t *testing.T) {
	_, err := ServerConfig(joinFile(samplesPath,
		conf.EptTest.ConfValidCaInvalid),
		true, conf.EptTest.ExportConfigKeys)
	if err == nil {
		t.Fatal(errPars)
	}
}

func TestPushConfig(t *testing.T) {
	fxt := data.NewTestFixture(t, db)
	defer fxt.Close()

	s := newTestSessSrv(0)
	defer s.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := PushConfig(ctx, conf.SessionServer.Config,
		logger, fxt.Product.ID, data.TestPassword,
		joinFile(samplesPath, conf.EptTest.ConfValidCaValid),
		joinFile(samplesPath, conf.EptTest.Ca),
		conf.EptTest.ExportConfigKeys, 100); err != nil {
		t.Fatal(err)
	}

	var prod data.Product

	if err := db.FindByPrimaryKeyTo(
		&prod, fxt.Product.ID); err != nil {
		t.Fatal(err)
	}

	var out map[string]string

	if err := json.Unmarshal(prod.Config, &out); err != nil {
		t.Fatal(err)
	}

	if !validParams(conf.EptTest.ExportConfigKeys, out) {
		t.Fatal(errPars)
	}
}

func TestPushConfigTimeout(t *testing.T) {
	fxt := data.NewTestFixture(t, db)
	defer fxt.Close()

	s := newTestSessSrv(time.Second)
	defer s.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := PushConfig(ctx, conf.SessionServer.Config,
		logger, fxt.Product.ID, data.TestPassword,
		joinFile(samplesPath, conf.EptTest.ConfValidCaValid),
		joinFile(samplesPath, conf.EptTest.Ca),
		conf.EptTest.ExportConfigKeys, 1); err != nil {
		t.Fatal(err)
	}

	var prod data.Product

	if err := db.FindByPrimaryKeyTo(
		&prod, fxt.Product.ID); err != nil {
		t.Fatal(err)
	}

	var out map[string]string

	if err := json.Unmarshal(prod.Config, &out); err != nil {
		t.Fatal(err)
	}

	if !validParams(conf.EptTest.ExportConfigKeys, out) {
		t.Fatal(errPars)
	}
}

func TestMain(m *testing.M) {
	conf.EptTest = newEptTestConfig()
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	conf.SessionServer = sesssrv.NewConfig()
	conf.SessionServerTest = newTestConfig()

	util.ReadTestConfig(&conf)

	logger = util.NewTestLogger(conf.Log)

	db = data.NewTestDB(conf.DB, logger)
	defer data.CloseDB(db)

	os.Exit(m.Run())
}
