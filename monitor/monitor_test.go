package monitor_test

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/monitor"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

var (
	db        *reform.DB
	ethClient *testEthereumClient
	mon       *monitor.Monitor
	pscABI    abi.ABI
	pscAddr   common.Address
)

type testContext struct {
	db        *reform.DB
	ethClient *testEthereumClient
	mon       *monitor.Monitor
}

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

func blockSettings(t *testing.T, fresh, limit,
	last, confirmations uint64) func() {
	settings := []*data.Setting{
		{
			Key:   data.SettingFreshBlocks,
			Name:  "fresh blocks",
			Value: fmt.Sprint(fresh),
		},
		{
			Key:   data.SettingBlockLimit,
			Name:  "block limit",
			Value: fmt.Sprint(limit),
		},
		{
			Key:   data.SettingLastProcessedBlock,
			Name:  "last scanned block",
			Value: fmt.Sprint(last),
		},
		{
			Key:   data.SettingMinConfirmations,
			Name:  "min confirmations",
			Value: fmt.Sprint(confirmations),
		},
	}

	data.InsertToTestDB(t, db, settings[0],
		settings[1], settings[2], settings[3])
	return func() {
		defer data.DeleteFromTestDB(t, db, settings[0],
			settings[1], settings[2], settings[3])
	}
}

func TestMain(m *testing.M) {
	var conf struct {
		DB  *data.DBConfig
		Log *log.FileConfig
		Job *job.Config
	}
	conf.DB = data.NewDBConfig()
	conf.Log = log.NewFileConfig()
	conf.Job = job.NewConfig()
	util.ReadTestConfig(&conf)
	db = data.NewTestDB(conf.DB)
	defer data.CloseDB(db)

	logger, err := log.NewStderrLogger(conf.Log.WriterConfig)
	if err != nil {
		panic(err)
	}

	p, err := abi.JSON(strings.NewReader(contract.PrivatixServiceContractABI))
	if err != nil {
		panic(err)
	}
	pscABI = p

	ethClient = &testEthereumClient{}
	queue := job.NewQueue(conf.Job, logger, db, nil)
	pscAddr = common.HexToAddress("0x1")
	mon, err = monitor.NewMonitor(&monitor.Config{}, ethClient, nil, db,
		logger, pscAddr, common.HexToAddress("0x2"), data.RoleAgent,
		queue)
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
