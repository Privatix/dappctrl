package somc_test

import (
	"os"
	"testing"

	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/agent/tor-somc"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

var (
	conf struct {
		DB        *data.DBConfig
		StderrLog *log.WriterConfig
	}
	db      *reform.DB
	handler *somc.Handler
)

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.StderrLog = log.NewWriterConfig()
	util.ReadTestConfig(&conf)

	db = data.NewTestDB(conf.DB)
	defer data.CloseDB(db)

	logger, err := log.NewStderrLogger(conf.StderrLog)
	if err != nil {
		panic(err)
	}

	handler = somc.NewHandler(db, logger)

	os.Exit(m.Run())
}
