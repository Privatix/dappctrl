package bc

import (
	"context"
	"math/big"
	"os"
	"testing"

	ethereum "github.com/ethereum/go-ethereum"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

type testEthereumClient struct {
	HeaderByNumberResult uint64
	FilterLogsResult     []ethtypes.Log
}

func (c *testEthereumClient) FilterLogs(context.Context,
	ethereum.FilterQuery) ([]ethtypes.Log, error) {
	return c.FilterLogsResult, nil
}

func (c *testEthereumClient) HeaderByNumber(ctx context.Context,
	number *big.Int) (*ethtypes.Header, error) {
	return &ethtypes.Header{
		Number: new(big.Int).SetUint64(c.HeaderByNumberResult),
	}, nil
}

func TestMain(m *testing.M) {
	var conf struct {
		DB  *data.DBConfig
		Log *log.WriterConfig
		Job *job.Config
	}
	conf.DB = data.NewDBConfig()
	conf.Log = log.NewWriterConfig()
	conf.Job = job.NewConfig()
	args := &util.TestArgs{
		Conf: &conf,
	}
	util.ReadTestArgs(args)
	db = data.NewTestDB(conf.DB)

	defer data.CloseDB(db)

	var err error
	logger, err = log.NewTestLogger(conf.Log, args.Verbose)
	if err != nil {
		panic(err)
	}

	queue = job.NewQueue(conf.Job, logger, db, nil)

	os.Exit(m.Run())
}
