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

	LogChannelTopUp = "LogChannelToppedUp"
)

var (
	conf *testConf

	logger *util.Logger
	db     *reform.DB

	pscAddr = common.HexToAddress(
		"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed")
	ptcAddr = common.HexToAddress(
		"0x0d825eb81b996c67a55f7da350b6e73bab3cb0ec")

	someAddress = common.HexToAddress(someAddressStr)
	someHash    = common.HexToHash(someHashStr)
)

type testConf struct {
	BlockMonitor *Config
	DB           *data.DBConfig
	Log          *util.LogConfig
	Job          *job.Config
}

type mockClient struct {
	logger  *util.Logger
	headers []ethtypes.Header
	logs    []ethtypes.Log
	number  uint64
}

type mockTicker struct {
	C chan time.Time
}

func newTestConf() *testConf {
	conf := new(testConf)
	conf.BlockMonitor = NewConfig()
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	conf.Job = job.NewConfig()
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

	addrBytes, err := data.ToBytes(acc.EthAddr)
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

func toHashes(t *testing.T, topics []interface{}) []common.Hash {
	hashes := make([]common.Hash, len(topics))
	if len(topics) > 0 {
		hashes[0] = common.HexToHash(topics[0].(string))
	}
	for i, topic := range topics[1:] {
		switch v := topic.(type) {
		case string:
			if bs, err := data.ToBytes(v); err == nil {
				hashes[i+1] = common.BytesToHash(bs)
			} else {
				t.Fatal(err)
			}
		case int:
			hashes[i+1] = common.BigToHash(big.NewInt(int64(v)))
		case common.Address:
			hashes[i+1] = v.Hash()
		case common.Hash:
			hashes[i+1] = v
		default:
			t.Fatalf("unsupported type %T as topic", topic)
		}
	}
	return hashes
}

func insertEvent(t *testing.T, db *reform.DB, blockNumber uint64,
	failures uint64, topics ...interface{}) *data.EthLog {
	el := &data.EthLog{
		ID:          util.NewUUID(),
		TxHash:      data.FromBytes(genRandData(32)),
		TxStatus:    txMinedStatus, // FIXME: is this field needed at all?
		BlockNumber: blockNumber,
		Addr:        data.FromBytes(pscAddr.Bytes()),
		Data:        data.FromBytes(genRandData(32)),
		Topics:      toHashes(t, topics),
		Failures:    failures,
	}
	if err := db.Insert(el); err != nil {
		t.Fatalf("failed to insert a log event"+
			" into db: %v", err)
	}

	return el
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
	c.logger.Debug("query: %v, filtered: %v", q, filtered)
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
		queue, client, pscAddr, ptcAddr)
	if err != nil {
		t.Fatal(err)
	}
	return mon, queue, client
}

func newErrorChecker(t *testing.T) chan error {
	ch := make(chan error)
	go checkErr(t, ch)
	return ch
}

func checkErr(t *testing.T, ch chan error) {
	for {
		err := <-ch
		t.Fatal(err)
	}
}

func setMaxRetryKey(t *testing.T) {
	setting := &data.Setting{Key: maxRetryKey, Value: "0"}
	data.InsertToTestDB(t, db, setting)
}

func TestMonitorLogCollect(t *testing.T) {
	defer cleanDB(t)

	mon, _, client := newTestObjects(t)

	errCh := newErrorChecker(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := newMockTicker()
	mon.start(ctx, 5, ticker.C, nil, errCh)

	_, agentAddress := insertNewAccount(t, db, agentPass)
	_, clientAddress := insertNewAccount(t, db, clientPass)

	eventAboutChannel := common.HexToHash(eth.EthDigestChannelCreated)
	eventAboutOffering := common.HexToHash(eth.EthOfferingCreated)
	eventAboutToken := common.HexToHash(eth.EthTokenApproval)

	var block uint64 = 10

	dataMap := make(map[string]bool)

	type logToInject struct {
		event  common.Hash
		agent  common.Address
		client common.Address
	}
	logsToInject := []logToInject{
		{eventAboutOffering, someAddress, someAddress}, // 1 match all offerings
		{someHash, someAddress, someAddress},           // 0 no match
		{someHash, agentAddress, someAddress},          // 1 match agent
		{someHash, someAddress, clientAddress},         // 0 match client, but not a client event
		// ----- 6 confirmations
		{eventAboutOffering, someAddress, someAddress},  // 1 match all offerings
		{eventAboutChannel, someAddress, someAddress},   // 0 no match
		{eventAboutToken, someAddress, someAddress},     // 0 no match
		{eventAboutChannel, agentAddress, someAddress},  // 1 match agent
		{eventAboutChannel, someAddress, clientAddress}, // 1 match client
		// ----- 2 confirmations
		{eventAboutOffering, agentAddress, someAddress}, // 1 match agent
		{eventAboutOffering, someAddress, someAddress},  // 1 match all offerings
		// ----- 0 confirmations
	}
	for _, contractAddr := range []common.Address{someAddress, pscAddr} {
		for _, log := range logsToInject {
			d := genRandData(32 * 5)
			dataMap[data.FromBytes(d)] = true
			client.injectEvent(&ethtypes.Log{
				Address:     contractAddr,
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

	cases := []struct {
		confirmations uint64
		freshnum      uint64
		lognum        int
	}{
		{6, 2, 2}, // freshnum = 2: will skip the first offering event
		{2, 0, 4}, // freshnum = 0: will include the second offering event
		{0, 2, 6},
	}

	var logs []*data.EthLog
	for _, c := range cases {
		setUint64Setting(t, db, minConfirmationsKey, c.confirmations)
		setUint64Setting(t, db, freshOfferingsKey, c.freshnum)
		ticker.tick()
		name := fmt.Sprintf("with %d confirmations and %d freshnum",
			c.confirmations, c.freshnum)
		logs = expectLogs(t, c.lognum, name, "")
	}

	for _, e := range logs {
		if !dataMap[e.Data] {
			t.Fatal("wrong data saved in a log entry")
		}
		delete(dataMap, e.Data)
	}
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
		data.FromBytes(someAddress.Bytes()),
		product.ID, template.ID,
	)
	offeringU := data.NewTestOffering(
		data.FromBytes(someAddress.Bytes()),
		product.ID, template.ID,
	)

	channel1 := data.NewTestChannel(
		acc1.EthAddr, data.FromBytes(someAddress.Bytes()),
		offering1.ID, 0, chanDepositVal, data.ChannelActive,
	)
	channel1.Block = 7
	channelX := data.NewTestChannel(
		data.FromBytes(someAddress.Bytes()), acc2.EthAddr,
		offeringX.ID, 0, chanDepositVal, data.ChannelActive,
	)
	channelX.Block = 8

	data.InsertToTestDB(t, db,
		product, template,
		offering1, offeringX, offeringU,
		channel1, channelX)

	return &testData{
		acc:      []*data.Account{acc1, acc2},
		addr:     []common.Address{addr1, addr2},
		product:  product,
		template: template,
		offering: []*data.Offering{offering1, offeringX, offeringU},
		channel:  []*data.Channel{channel1, channelX},
	}
}

func scheduleTest(t *testing.T, td *testData, queue *mockQueue,
	ticker *mockTicker, mon *Monitor) {
	var blockNum uint64
	nextBlock := func() uint64 {
		blockNum++
		return blockNum
	}

	insertEvent(t, db, nextBlock(), 0,
		eth.EthTokenApproval,
		td.addr[0],
		pscAddr,
		123)

	queue.expect(data.JobPreAccountAddBalance, func(j *data.Job) bool {
		return j.Type == data.JobPreAccountAddBalance
	})

	insertEvent(t, db, nextBlock(), 0,
		eth.EthTokenTransfer,
		td.addr[0],
		someAddress,
		123)

	queue.expect(data.JobAfterAccountAddBalance, func(j *data.Job) bool {
		return j.Type == data.JobAfterAccountAddBalance
	})

	insertEvent(t, db, nextBlock(), 0,
		eth.EthTokenTransfer,
		someAddress,
		td.addr[0],
		123)

	queue.expect(data.JobAfterAccountAddBalance, func(j *data.Job) bool {
		return j.Type == data.JobAfterAccountAddBalance
	})

	insertEvent(t, db, nextBlock(), 0,
		eth.EthOfferingCreated,
		td.addr[0],          // agent
		td.offering[0].Hash, // offering hash
		minDepositVal,       // min deposit
	)
	queue.expect(data.JobAgentAfterOfferingMsgBCPublish, func(j *data.Job) bool {
		return j.Type == data.JobAgentAfterOfferingMsgBCPublish
	})
	// offering events containing agent address should be ignored

// TODO: uncomment, when client job handler are completed
/*
	insertEvent(t, db, nextBlock(), 0,
		eth.EthOfferingCreated,
		someAddress,         // agent
		td.offering[2].Hash, // offering hash
		minDepositVal,       // min deposit
	)
	queue.expect(unrelatedOfferingCreated, func(j *data.Job) bool {
		return j.Type == data.JobClientAfterOfferingMsgBCPublish
	})

	insertEvent(t, db, nextBlock(), 0,
		eth.EthOfferingPoppedUp,
		someAddress,         // agent
		td.offering[2].Hash, // offering hash
	)
	queue.expect(clientOfferingPoppedUp, func(j *data.Job) bool {
		return j.Type == data.JobClientAfterOfferingMsgBCPublish
	})

	// Tick here on purpose, so that not all events are ignored because
	// the offering's been deleted.
	ticker.tick()
	queue.awaitCompletion(time.Second)

	insertEvent(t, db, nextBlock(), 0,
		eth.EthOfferingDeleted,
		someAddress,         // agent
		td.offering[2].Hash, // offering hash
	)
	// should ignore the deletion event
*/
	insertEvent(t, db, nextBlock(), 0,
		eth.EthOfferingPoppedUp,
		someAddress,         // agent
		td.offering[2].Hash, // offering hash
	)
	// should ignore the creation event after deleting

	insertEvent(t, db, nextBlock(), 0,
		eth.EthDigestChannelCreated,
		td.addr[0],          // agent
		someAddress,         // client
		td.offering[0].Hash, // offering
	)
	queue.expect(agentAfterChannelCreated, func(j *data.Job) bool {
		return j.Type == data.JobAgentAfterChannelCreate
	})
// TODO: uncomment, when client job handler are completed
/*
	el := insertEvent(t, db, nextBlock(), 0,
		eth.EthDigestChannelToppedUp,
		td.addr[0],          // agent
		someAddress,         // client
		td.offering[1].Hash, // offering
	)
	bs, err := mon.pscABI.Events[LogChannelTopUp].Inputs.NonIndexed().Pack(
		uint32(td.channel[1].Block),
		new(big.Int),
	)
	if err != nil {
		t.Fatal(err)
	}
	el.Data = data.FromBytes(bs)
	if err := db.Save(el); err != nil {
		t.Fatal(err)
	}
	// channel does not exist, thus event ignored

	el = insertEvent(t, db, nextBlock(), 0,
		eth.EthDigestChannelToppedUp,
		someAddress,         // agent
		td.addr[1],          // client
		td.offering[1].Hash, // offering
	)

	bs, err = mon.pscABI.Events[LogChannelTopUp].Inputs.NonIndexed().Pack(
		uint32(td.channel[1].Block),
		new(big.Int),
	)
	if err != nil {
		t.Fatal(err)
	}
	el.Data = data.FromBytes(bs)
	if err := db.Save(el); err != nil {
		t.Fatal(err)
	}
	queue.expect(clientAfterChannelTopUp, func(j *data.Job) bool {
		return j.Type == data.JobClientAfterChannelTopUp
	})

	ticker.tick()
	queue.awaitCompletion(time.Second)
*/
}

func TestMonitorSchedule(t *testing.T) {
	defer cleanDB(t)

	setMaxRetryKey(t)

	mon, queue, _ := newTestObjects(t)

	errCh := newErrorChecker(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := newMockTicker()
	mon.start(ctx, conf.BlockMonitor.Timeout, nil, ticker.C, errCh)

	td := generateTestData(t)

	scheduleTest(t, td, queue, ticker, mon)
}

// TestMain reads config and run tests.
func TestMain(m *testing.M) {
	conf = newTestConf()
	util.ReadTestConfig(&conf)

	logger = util.NewTestLogger(conf.Log)
	db = data.NewTestDB(conf.DB, logger)
	defer data.CloseDB(db)

	os.Exit(m.Run())
}
