// +build !nomonitortest

package monitor

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/data"
)

type mockClient struct {
	logger  *util.Logger
	headers []ethtypes.Header
	logs    []ethtypes.Log
	number  uint64
}

func addressIsAmong(x *common.Address, addresses []common.Address) bool {
	for _, a := range addresses {
		if *x == a {
			return true
		}
	}
	return false
}

func hashIsAmong(x *common.Hash, hashes []common.Hash) bool {
	for _, h := range hashes {
		if *x == h {
			return true
		}
	}
	return false
}

func eventSatisfiesFilter(e *ethtypes.Log, q ethereum.FilterQuery) bool {
	if q.FromBlock != nil && e.BlockNumber < q.FromBlock.Uint64() {
		return false
	}

	if q.ToBlock != nil && e.BlockNumber > q.ToBlock.Uint64() {
		return false
	}

	if len(q.Addresses) > 0 && !addressIsAmong(&e.Address, q.Addresses) {
		return false
	}

	for i, hashes := range q.Topics {
		if len(hashes) > 0 {
			if i >= len(e.Topics) {
				return false
			}
			if !hashIsAmong(&e.Topics[i], hashes) {
				return false
			}
		}
	}

	return true
}

func (c *mockClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]ethtypes.Log, error) {
	var filtered []ethtypes.Log
	for _, e := range c.logs {
		if eventSatisfiesFilter(&e, q) {
			filtered = append(filtered, e)
		}
	}
	c.logger.Debug("query: %v, filtered: %v", q, filtered)
	return filtered, nil
}

// HeaderByNumber returns a minimal header for testing. It only supports calls where number is nil.
// Moreover, only the Number field in the returned header is valid.
func (c *mockClient) HeaderByNumber(ctx context.Context, number *big.Int) (*ethtypes.Header, error) {
	if number != nil {
		return nil, fmt.Errorf("mock HeaderByNumber() only supports nil as 'number'")
	}
	return &ethtypes.Header{
		Number: new(big.Int).SetUint64(c.number),
	}, nil
}

func (c *mockClient) injectEvent(e *ethtypes.Log) {
	c.logs = append(c.logs, *e)
	if c.number < e.BlockNumber {
		c.number = e.BlockNumber
	}
}

type mockTicker struct {
	C chan time.Time
}

func newMockTicker() *mockTicker {
	return &mockTicker{C: make(chan time.Time, 1)}
}

func (t *mockTicker) tick() {
	select {
		case t.C <- time.Now():
		default:
	}
}

var (
	logger     *util.Logger
	db         *reform.DB
	client     mockClient

	pscAddr    = common.HexToAddress("0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed")
)

// TestMain reads config and run tests.
func TestMain(m *testing.M) {
	var conf struct {
		DB              *data.DBConfig
		Log             *util.LogConfig
		Eth struct {
			GethURL       string
			TruffleAPIURL string
		}
	}
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	util.ReadTestConfig(&conf)

	logger = util.NewTestLogger(conf.Log)
	client.logger = logger
	db = data.NewTestDB(conf.DB, logger)
	defer data.CloseDB(db)

	os.Exit(m.Run())
}

func cleanDB(t *testing.T) {
	data.CleanTestDB(t, db)
}

func expectLogs(t *testing.T, expected int, errMsg, tail string, args ...interface{}) []*data.LogEntry {
	var (
		actual int
	)
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		structs, err := db.SelectAllFrom(data.LogEntryTable, tail, args...)
		if err != nil {
			t.Fatalf("failed to select log entries: %v", err)
		}
		actual = len(structs)
		if actual == expected {
			logs := make([]*data.LogEntry, actual)
			for li, s := range structs {
				logs[li] = s.(*data.LogEntry)
			}
			return logs
		}
	}
	t.Fatalf("%s: wrong number of log entries collected: got %d, expected %d", errMsg, actual, expected)
	return nil
}

func insertNewAccount(t *testing.T, db *reform.DB, auth string) (*data.Account, common.Address) {
	acc := data.NewTestAccount(auth)
	data.InsertToTestDB(t, db, acc)

	addrBytes, _ := data.ToBytes(acc.EthAddr)
	addr := common.BytesToAddress(addrBytes)
	return acc, addr
}

func genRandData(length int) []byte {
	randbytes := make([]byte, length)
	rand.Read(randbytes)
	return randbytes
}

func setMinConfirmations(t *testing.T, db *reform.DB, value uint64) {
	setting := data.Setting{
		Key: "eth.min.confirmations",
		Value: strconv.FormatUint(value, 10),
		Name: "eth.min.confirmations",
	}
	if err := db.Save(&setting); err != nil {
		t.Fatalf("failed to save min confirmtions setting: %v", err)
	}
}

func TestMonitorLogCollect(t *testing.T) {
	defer cleanDB(t)

	mon := NewMonitor(logger, db, &client, pscAddr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := newMockTicker()
	if err := mon.start(ctx, ticker.C, nil); err != nil {
		panic(err)
	}

	_, agentAddress := insertNewAccount(t, db, "agentpass")
	_, clientAddress := insertNewAccount(t, db, "clientpass")

	someAddress := common.HexToAddress("0xdeadbeef")
	someHash := common.HexToHash("0xc0ffee")

	eventForClient := common.HexToHash(eth.EthOfferingEndpoint)
	eventForAgent := common.HexToHash(eth.EthOfferingCreated)

	var block uint64 = 10

	datamap := make(map[string]bool)
	logsToInject := []struct{
		event common.Hash
		agent common.Address
		client common.Address
	}{
		{someHash, someAddress, someAddress}, // 0 no match
		{someHash, agentAddress, someAddress}, // 1 match agent
		{someHash, someAddress, clientAddress}, // 0 match client, but not a client event
		// ----- 6 confirmations
		{eventForAgent, someAddress, someAddress}, // 0 no match
		{eventForAgent, agentAddress, someAddress}, // 1 match agent
		{eventForAgent, someAddress, clientAddress}, // 0 match client, but not a client event
		{eventForClient, someAddress, someAddress}, // 0 no match
		// ----- 2 confirmations
		{eventForClient, agentAddress, someAddress}, // 1 match agent
		{eventForClient, someAddress, clientAddress}, // 1 match client
		// ----- 0 confirmations
	}
	for _, contractAddr := range []common.Address{someAddress, pscAddr} {
		for _, log := range logsToInject {
			d := genRandData(32 * 5)
			datamap[data.FromBytes(d)] = true
			client.injectEvent(&ethtypes.Log{
				Address: contractAddr,
				BlockNumber: block,
				Topics: []common.Hash{
					log.event,
					log.agent.Hash(),
					log.client.Hash(),
				},
				Data: d,
			})
			block++
		}
	}

	cases := []struct{
		confirmations uint64
		lognum int
	}{
		{6, 1},
		{2, 2},
		{0, 4},
	}

	var logs []*data.LogEntry
	for _, c := range cases {
		setMinConfirmations(t, db, c.confirmations)
		ticker.tick()
		name := fmt.Sprintf("with %d confirmations", c.confirmations)
		logs = expectLogs(t, c.lognum, name, "")
	}

	for _, e := range logs {
		if !datamap[e.Data] {
			t.Fatalf("wrong data saved in a log entry")
		}
		delete(datamap, e.Data)
	}
}
