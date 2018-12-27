package somcsrv_test

import (
	"os"
	"testing"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/agent/somcsrv"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

var (
	conf struct {
		DB  *data.DBConfig
		Log *log.WriterConfig
	}
	db      *reform.DB
	handler *somcsrv.Handler
)

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.Log = log.NewWriterConfig()
	args := &util.TestArgs{
		Conf: &conf,
	}
	util.ReadTestArgs(args)

	logger, err := log.NewTestLogger(conf.Log, args.Verbose)
	if err != nil {
		panic(err)
	}

	db = data.NewTestDB(conf.DB)
	defer data.CloseDB(db)

	handler = somcsrv.NewHandler(db, logger)

	os.Exit(m.Run())
}
