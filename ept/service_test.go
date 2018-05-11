package ept

import (
	"os"
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/pay"
	"github.com/privatix/dappctrl/util"
)

var (
	conf struct {
		DB        *data.DBConfig
		Log       *util.LogConfig
		PayServer *pay.Config
		EptTest   *eptTestConfig
	}

	testDB *reform.DB

	timeout = time.Second * 30
)

type eptTestConfig struct {
	Channel             string
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
}

func newEptTestConfig() *eptTestConfig {
	return &eptTestConfig{}
}

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	conf.PayServer = &pay.Config{}
	conf.EptTest = newEptTestConfig()

	util.ReadTestConfig(&conf)

	logger := util.NewTestLogger(conf.Log)

	testDB = data.NewTestDB(conf.DB, logger)

	defer data.CloseDB(testDB)

	os.Exit(m.Run())

}

func TestServiceEndpointMessage(t *testing.T) {
	s := New(testDB, conf.PayServer)

	_, err := s.EndpointMessage(conf.EptTest.Channel, timeout)
	if err != nil {
		t.Fatal(err)
	}
}
