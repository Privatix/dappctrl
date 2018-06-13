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
	errPars = "incorrect parsing test"
	errPush = "incorrect pushing test"

	samplesPath = "samples"
)

var (
	conf struct {
		DB                *data.DBConfig
		EptTest           *eptTestConfig
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
	EmptyConfig       string
	ValidFormatConfig map[string]string
}

type cliConfTest struct {
	Login string
	Pass  string
}

type eptTestConfig struct {
	Ca                  string
	ConfValidCaValid    string
	ConfInvalid         string
	ConfValidCaInvalid  string
	ConfValidCaEmpty    string
	ConfValidCaNotExist string
	VPNConfig           cliConfTest
	SessSrvStartTimeout []int64
	ExportConfigKeys    []string
	RetrySec            []int64
	ValidHost           []string
}

func newTestSessSrv(timeout time.Duration) *testSessSrv {
	srv := sesssrv.NewServer(conf.SessionServer, logger, db)
	go func() {
		time.Sleep(timeout)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("failed to serve session requests: %s", err)
		}
	}()

	time.Sleep(time.Duration(
		conf.SessionServerTest.ServerStartupDelay) * time.Millisecond)

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

func validParams(out map[string]string) bool {
	for _, key := range conf.EptTest.ExportConfigKeys {
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

func testContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

func config(t *testing.T, id string) map[string]string {
	var prod data.Product

	if err := db.FindByPrimaryKeyTo(
		&prod, id); err != nil {
		t.Fatal(err)
	}

	var out map[string]string

	if err := json.Unmarshal(prod.Config, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func pushConfig(ctx context.Context, productID string, keys []string,
	retrySec int64) error {
	req := NewPushConfigReq(productID, data.TestPassword,
		joinFile(samplesPath, conf.EptTest.ConfValidCaValid),
		joinFile(samplesPath, conf.EptTest.Ca), keys, retrySec)

	return PushConfig(ctx, conf.SessionServer.Config, logger, req)
}

func srvConfig(t *testing.T) map[string]string {
	out, err := ServerConfig(joinFile(samplesPath,
		conf.EptTest.ConfValidCaValid),
		true, conf.EptTest.ExportConfigKeys)
	if err != nil {
		t.Fatal(err.Error())
	}
	return out
}

func TestParsingValidConfig(t *testing.T) {
	if !validParams(srvConfig(t)) {
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

	s := newTestSessSrv(time.Duration(conf.EptTest.SessSrvStartTimeout[0]))
	defer s.Close()

	ctx, cancel := testContext()
	defer cancel()

	if err := pushConfig(ctx, fxt.Product.ID,
		conf.EptTest.ExportConfigKeys,
		conf.EptTest.RetrySec[1]); err != nil {
		t.Fatal(err)
	}

	c := config(t, fxt.Product.ID)

	if !validParams(c) {
		t.Fatal(errPars)
	}
}

func TestPushConfigTimeout(t *testing.T) {
	fxt := data.NewTestFixture(t, db)
	defer fxt.Close()

	s := newTestSessSrv(time.Second *
		time.Duration(conf.EptTest.SessSrvStartTimeout[1]))
	defer s.Close()

	ctx, cancel := testContext()
	defer cancel()

	if err := pushConfig(ctx, fxt.Product.ID,
		conf.EptTest.ExportConfigKeys,
		conf.EptTest.RetrySec[2]); err != nil {
		t.Fatal(err)
	}

	c := config(t, fxt.Product.ID)

	if !validParams(c) {
		t.Fatal(errPars)
	}
}

func TestPushConfigCancel(t *testing.T) {
	fxt := data.NewTestFixture(t, db)
	defer fxt.Close()

	ctx, cancel := testContext()
	go func() {
		time.Sleep(time.Second *
			time.Duration(conf.EptTest.SessSrvStartTimeout[2]))
		cancel()
	}()

	err := pushConfig(ctx, fxt.Product.ID, conf.EptTest.ExportConfigKeys,
		conf.EptTest.RetrySec[1])
	if err == nil || err != ErrCancel {
		t.Fatal(errPush)
	}

	c := config(t, fxt.Product.ID)

	if len(c) != 0 {
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
