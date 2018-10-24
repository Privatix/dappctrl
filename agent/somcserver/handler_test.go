package somcserver_test

import (
	"os"
	"testing"

	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/agent/somcserver"
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
	handler *somcserver.Handler
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

	handler = somcserver.NewHandler(db, logger)

	os.Exit(m.Run())
}
