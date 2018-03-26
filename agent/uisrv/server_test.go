// +build !noagentuisrvtest

package uisrv

import (
	"os"
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

// used throughout all tests in the package.
var (
	testServer *Server
)

func TestMain(m *testing.M) {
	var conf struct {
		AgentServer *Config
		DB          *data.DBConfig
		Log         *util.LogConfig
	}
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	util.ReadTestConfig(&conf)
	logger := util.NewTestLogger(conf.Log)
	db := data.NewTestDB(conf.DB, logger)
	defer data.CloseDB(db)
	testServer = NewServer(conf.AgentServer, logger, db)

	os.Exit(m.Run())
}
