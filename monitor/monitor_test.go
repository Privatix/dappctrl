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

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

const (
	agentPass      = "agentpass"
	clientPass     = "clientpass"
	someAddressStr = "0xdeadbeef"
	someHashStr    = "0xc0ffee"

	minDepositVal  = 123
	chanDepositVal = 100

	unrelatedOfferingCreated = "unrelated offering created"
	clientOfferingPoppedUp   = "client offering popped up"
	agentAfterChannelCreated = "agent after channel created"
	clientAfterChannelTopUp  = "client after channel topup"

	txHash = "d8de4d04f002759b9153bb15a8e81a86700609e69c1a28f7eaa11643b754679d"
)

var (
	conf *testConf

	logger log.Logger
	db     *reform.DB

	pscAddr = common.HexToAddress(
		"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed")
	ptcAddr = common.HexToAddress(
		"0x0d825eb81b996c67a55f7da350b6e73bab3cb0ec")

	someAddress = common.HexToAddress(someAddressStr)
	someHash    = common.HexToHash(someHashStr)

	blockNum uint64
)

type testConf struct {
	BlockMonitor *Config
	DB           *data.DBConfig
	Log          *log.FileConfig
	Job          *job.Config
	Eth          *eth.Config
}

type mockClient struct {
	logger  log.Logger
	headers []ethtypes.Header
	logs    []ethtypes.Log
	number  uint64
}

type mockTicker struct {
	C chan time.Time
}

func nextBlock() uint64 {
	blockNum++
	return blockNum
}

func newTestConf() *testConf {
	conf := new(testConf)
	conf.BlockMonitor = NewConfig()
	conf.DB = data.NewDBConfig()
	conf.Log = log.NewFileConfig()
	conf.Job = job.NewConfig()
	conf.Eth = eth.NewConfig()
	return conf
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

func newMockTicker() *mockTicker {
	return &mockTicker{C: make(chan time.Time, 1)}
}

func (t *mockTicker) tick() {
	select {
	case t.C <- time.Now():
	default:
	}
}

func cleanDB(t *testing.T) {
	data.CleanTestDB(t, db)
}

func expectLogs(t *testing.T, expected int, errMsg, tail string,
	args ...interface{}) []*data.EthLog {
	var (
		actual int
	)
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		structs, err := db.SelectAllFrom(data.EthLogTable,
			tail, args...)
		if err != nil {
			t.Fatalf("failed to select log entries: %v", err)
		}
		actual = len(structs)
		if actual == expected {
			logs := make([]*data.EthLog, actual)
			for li, s := range structs {
				logs[li] = s.(*data.EthLog)
			}
			return logs
		}
	}
	t.Fatalf("%s: wrong number of log entries collected:"+
		" got %d, expected %d", errMsg, actual, expected)
	return nil
}

func insertNewAccount(t *testing.T, db *reform.DB,
	auth string) (*data.Account, common.Address) {
	acc := data.NewTestAccount(auth)
	data.InsertToTestDB(t, db, acc)

	addrBytes, err := data.HexToBytes(acc.EthAddr)
	if err != nil {
		t.Fatal(err)
	}

	addr := common.BytesToAddress(addrBytes)
	return acc, addr
}

func genRandData(length int) []byte {
	randbytes := make([]byte, length)
	rand.Read(randbytes)
	return randbytes
}

func setUint64Setting(t *testing.T, db *reform.DB,
	key string, value uint64) {
	setting := data.Setting{
		Key:   key,
		Value: strconv.FormatUint(value, 10),
		Name:  key,
	}
	if err := db.Save(&setting); err != nil {
		t.Fatalf("failed to save min confirmtions"+
			" setting: %v", err)
	}
}

type expectation struct {
	condition func(j *data.Job) bool
	comment   string
}

type mockQueue struct {
	t            *testing.T
	db           *reform.DB
	expectations []expectation
}

func newMockQueue(t *testing.T, db *reform.DB) *mockQueue {
	return &mockQueue{t: t, db: db}
}

func (mq *mockQueue) Add(j *data.Job) error {
	if len(mq.expectations) == 0 {
		mq.t.Fatalf("unexpected job added, expected none, got %#v", *j)
	}
	ex := mq.expectations[0]
	mq.expectations = mq.expectations[1:]
	if !ex.condition(j) {
		mq.t.Fatalf("unexpected job added, expected %s, got %#v",
			ex.comment, *j)
	}
	j.ID = util.NewUUID()
	j.Status = data.JobActive
	data.InsertToTestDB(mq.t, mq.db, j)
	return nil
}

func (mq *mockQueue) expect(comment string, condition func(j *data.Job) bool) {
	mq.expectations = append(mq.expectations,
		expectation{condition, comment})
}

func (mq *mockQueue) awaitCompletion(timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if len(mq.expectations) == 0 {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	mq.t.Fatalf("not all expected jobs scheduled: %d left",
		len(mq.expectations))
}

func newMockClient() *mockClient {
	client := &mockClient{}
	client.logger = logger
	return client
}

func (c *mockClient) FilterLogs(ctx context.Context,
	q ethereum.FilterQuery) ([]ethtypes.Log, error) {
	var filtered []ethtypes.Log
	for _, e := range c.logs {
		if eventSatisfiesFilter(&e, q) {
			filtered = append(filtered, e)
		}
	}
	c.logger.Debug(fmt.Sprintf("query: %v, filtered: %v", q, filtered))
	return filtered, nil
}

// HeaderByNumber returns a minimal header for testing.
// It only supports calls where number is nil.
// Moreover, only the Number field in the returned header is valid.
func (c *mockClient) HeaderByNumber(ctx context.Context,
	number *big.Int) (*ethtypes.Header, error) {
	if number != nil {
		return nil, fmt.Errorf("mock HeaderByNumber()" +
			" only supports nil as 'number'")
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

func newTestObjects(t *testing.T) (*Monitor, *mockQueue, *mockClient) {
	queue := newMockQueue(t, db)

	client := newMockClient()

	mon, err := NewMonitor(conf.BlockMonitor, logger, db,
		queue, conf.Eth, pscAddr, ptcAddr, client, "", nil)
	if err != nil {
		t.Fatal(err)
	}

	return mon, queue, client
}

func setMaxRetryKey(t *testing.T) {
	setting := &data.Setting{Key: maxRetryKey, Value: "0"}
	data.InsertToTestDB(t, db, setting)
}

type testData struct {
	acc      []*data.Account
	addr     []common.Address
	product  *data.Product
	template *data.Template
	offering []*data.Offering
	channel  []*data.Channel
}

func generateTestData(t *testing.T) *testData {
	acc1, addr1 := insertNewAccount(t, db, clientPass)
	acc2, addr2 := insertNewAccount(t, db, clientPass)

	product := data.NewTestProduct()
	template := data.NewTestTemplate(data.TemplateOffer)

	offering1 := data.NewTestOffering(
		acc1.EthAddr, product.ID, template.ID)
	offeringX := data.NewTestOffering(
		data.HexFromBytes(someAddress.Bytes()),
		product.ID, template.ID,
	)
	offeringU := data.NewTestOffering(
		data.HexFromBytes(someAddress.Bytes()),
		product.ID, template.ID,
	)

	channel1 := data.NewTestChannel(
		acc1.EthAddr, data.HexFromBytes(someAddress.Bytes()),
		offering1.ID, 0, chanDepositVal, data.ChannelActive,
	)
	channel1.Block = 7
	channelX := data.NewTestChannel(
		data.HexFromBytes(someAddress.Bytes()), acc2.EthAddr,
		offeringX.ID, 0, chanDepositVal, data.ChannelActive,
	)
	channelX.Block = 8

	createChanneljob := data.NewTestJob(data.JobClientPreChannelCreate,
		data.JobUser, data.JobOffering)
	createChanneljob.RelatedID = channelX.ID
	createChanneljob.Status = data.JobDone

	tx := &data.EthTx{
		ID:          util.NewUUID(),
		Hash:        txHash,
		Method:      "CreateChannel",
		Status:      data.TxSent,
		JobID:       &createChanneljob.ID,
		Issued:      time.Now(),
		RelatedType: data.JobChannel,
		RelatedID:   channelX.ID,
		Gas:         1,
		GasPrice:    uint64(1),
		TxRaw:       []byte("{}"),
	}

	data.InsertToTestDB(t, db,
		product, template,
		offering1, offeringX, offeringU,
		channel1, channelX, createChanneljob, tx)

	return &testData{
		acc:      []*data.Account{acc1, acc2},
		addr:     []common.Address{addr1, addr2},
		product:  product,
		template: template,
		offering: []*data.Offering{offering1, offeringX, offeringU},
		channel:  []*data.Channel{channel1, channelX},
	}
}

// TestMain reads config and run tests.
func TestMain(m *testing.M) {
	conf = newTestConf()
	util.ReadTestConfig(&conf)

	l, err := log.NewStderrLogger(conf.Log)
	if err != nil {
		panic(err)
	}

	logger = l

	db = data.NewTestDB(conf.DB)
	defer data.CloseDB(db)

	os.Exit(m.Run())
}
