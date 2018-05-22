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
	logger *util.Logger
	db     *reform.DB
	client mockClient

	pscAddr = common.HexToAddress("0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed")
)

// TestMain reads config and run tests.
func TestMain(m *testing.M) {
	var conf struct {
		DB  *data.DBConfig
		Log *util.LogConfig
		Job *job.Config
		Eth struct {
			GethURL       string
			TruffleAPIURL string
		}
	}
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	conf.Job = job.NewConfig()
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

func expectLogs(t *testing.T, expected int, errMsg, tail string, args ...interface{}) []*data.EthLog {
	var (
		actual int
	)
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		structs, err := db.SelectAllFrom(data.EthLogTable, tail, args...)
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

func setUint64Setting(t *testing.T, db *reform.DB, key string, value uint64) {
	setting := data.Setting{
		Key:   key,
		Value: strconv.FormatUint(value, 10),
		Name:  key,
	}
	if err := db.Save(&setting); err != nil {
		t.Fatalf("failed to save min confirmtions setting: %v", err)
	}
}

func TestMonitorLogCollect(t *testing.T) {
	defer cleanDB(t)

	mon := NewMonitor(logger, db, nil, &client, pscAddr)

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

	eventAboutChannel := common.HexToHash(eth.EthDigestChannelCreated)
	eventAboutOffering := common.HexToHash(eth.EthOfferingCreated)

	var block uint64 = 10

	datamap := make(map[string]bool)
	logsToInject := []struct {
		event  common.Hash
		agent  common.Address
		client common.Address
	}{
		{eventAboutOffering, someAddress, someAddress}, // 1 match all offerings
		{someHash, someAddress, someAddress},           // 0 no match
		{someHash, agentAddress, someAddress},          // 1 match agent
		{someHash, someAddress, clientAddress},         // 0 match client, but not a client event
		// ----- 6 confirmations
		{eventAboutOffering, someAddress, someAddress},  // 1 match all offerings
		{eventAboutChannel, someAddress, someAddress},   // 0 no match
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
			datamap[data.FromBytes(d)] = true
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
		{6, 2, 1}, // freshnum = 2: will skip the first offering event
		{2, 0, 4}, // freshnum = 0: will include the second offering event
		{0, 2, 6},
	}

	var logs []*data.EthLog
	for _, c := range cases {
		setUint64Setting(t, db, minConfirmationsKey, c.confirmations)
		setUint64Setting(t, db, freshOfferingsKey, c.freshnum)
		ticker.tick()
		name := fmt.Sprintf("with %d confirmations and %d freshnum", c.confirmations, c.freshnum)
		logs = expectLogs(t, c.lognum, name, "")
	}

	for _, e := range logs {
		if !datamap[e.Data] {
			t.Fatalf("wrong data saved in a log entry")
		}
		delete(datamap, e.Data)
	}
}

func insertEvent(db *reform.DB, blockNumber uint64, topics []string, failures int) {
	el := &data.EthLog{
		ID:          util.NewUUID(),
		TxHash:      data.FromBytes(genRandData(32)),
		TxStatus:    "mined", // FIXME: is this field needed at all?
		BlockNumber: blockNumber,
		Addr:        data.FromBytes(pscAddr.Bytes()),
		Data:        data.FromBytes(genRandData(32)),
		Topics:      topics,
	}
	if err := db.Insert(el); err != nil {
		panic(fmt.Errorf("failed to insert a log event into db: %v", err))
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

func (mq *mockQueue) Add(j *data.Job) error {
	if len(mq.expectations) == 0 {
		mq.t.Fatalf("unexpected job added, expected none")
	}
	ex := mq.expectations[0]
	mq.expectations = mq.expectations[1:]
	if !ex.condition(j) {
		mq.t.Fatalf("unexpected job added, expected %s, got %#v", ex.comment, *j)
	}
	j.ID = util.NewUUID()
	j.Status = data.JobActive
	data.InsertToTestDB(mq.t, mq.db, j)
	return nil
}

func (mq *mockQueue) expect(comment string, condition func(j *data.Job) bool) {
	mq.expectations = append(mq.expectations, expectation{condition, comment})
}

func (mq *mockQueue) awaitCompletion(timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if len(mq.expectations) == 0 {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	mq.t.Fatal("not all expected jobs scheduled")
}

func b64toHashHex(b64 string) string {
	bs, _ := data.ToBytes(b64)
	return common.BytesToHash(bs).Hex()
}

func TestMonitorLogSchedule(t *testing.T) {
	defer cleanDB(t)

	queue := &mockQueue{t: t, db: db}
	mon := NewMonitor(logger, db, queue, nil, pscAddr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := newMockTicker()
	if err := mon.start(ctx, nil, ticker.C); err != nil {
		panic(err)
	}

	var blockNum uint64
	nextBlock := func() uint64 {
		blockNum++
		return blockNum
	}

	someAddress := common.HexToAddress("0xdeadbeef")
	//someHash := common.HexToHash("0xc0ffee")

	acc1, addr1 := insertNewAccount(t, db, "clientpass")
	acc2, addr2 := insertNewAccount(t, db, "clientpass")

	product := data.NewTestProduct()
	template := data.NewTestTemplate(data.TemplateOffer)

	offering1 := data.NewTestOffering(acc1.EthAddr, product.ID, template.ID)
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
		offering1.ID, 0, 100, data.ChannelActive,
	)
	channelX := data.NewTestChannel(
		data.FromBytes(someAddress.Bytes()), acc2.EthAddr,
		offeringX.ID, 0, 100, data.ChannelActive,
	)

	data.InsertToTestDB(t, db,
		product, template,
		offering1, offeringX, offeringU,
		channel1, channelX)

	insertEvent(db, nextBlock(), []string{
		"0x" + eth.EthOfferingCreated,
		addr1.Hex(),                  // agent
		b64toHashHex(offering1.Hash), // offering hash
		"0x123", // min deposit
	}, 0)
	// offering events containing agent address should be ignored

	insertEvent(db, nextBlock(), []string{
		"0x" + eth.EthOfferingCreated,
		someAddress.Hex(),            // agent
		b64toHashHex(offeringU.Hash), // offering hash
		"0x123", // min deposit
	}, 0)
	queue.expect("unrelated offering created", func(j *data.Job) bool {
		return j.Type == data.JobClientAfterOfferingMsgBCPublish
	})

	insertEvent(db, nextBlock(), []string{
		"0x" + eth.EthOfferingPoppedUp,
		someAddress.Hex(),            // agent
		b64toHashHex(offeringU.Hash), // offering hash
	}, 0)
	queue.expect("client offering popped up", func(j *data.Job) bool {
		return j.Type == data.JobClientAfterOfferingMsgBCPublish
	})

	// Tick here on purpose, so that not all events are ignored because
	// the offering's been deleted.
	ticker.tick()
	queue.awaitCompletion(time.Second)

	insertEvent(db, nextBlock(), []string{
		"0x" + eth.EthOfferingDeleted,
		someAddress.Hex(),            // agent
		b64toHashHex(offeringU.Hash), // offering hash
	}, 0)
	// should ignore the deletion event

	insertEvent(db, nextBlock(), []string{
		"0x" + eth.EthOfferingPoppedUp,
		someAddress.Hex(),            // agent
		b64toHashHex(offeringU.Hash), // offering hash
	}, 0)
	// should ignore the creation event after deleting

	insertEvent(db, nextBlock(), []string{
		"0x" + eth.EthDigestChannelCreated,
		addr1.Hex(),                  // agent
		someAddress.Hex(),            // client
		b64toHashHex(offering1.Hash), // offering
	}, 0)
	queue.expect("agent after channel created", func(j *data.Job) bool {
		return j.Type == data.JobAgentAfterChannelCreate
	})

	insertEvent(db, nextBlock(), []string{
		"0x" + eth.EthDigestChannelToppedUp,
		addr1.Hex(),                  // agent
		someAddress.Hex(),            // client
		b64toHashHex(offeringX.Hash), // offering
	}, 0)
	// channel does not exist, thus event ignored

	insertEvent(db, nextBlock(), []string{
		"0x" + eth.EthDigestChannelToppedUp,
		someAddress.Hex(),            // agent
		addr2.Hex(),                  // client
		b64toHashHex(offeringX.Hash), // offering
	}, 0)
	queue.expect("client after channel topup", func(j *data.Job) bool {
		return j.Type == data.JobClientAfterChannelTopUp
	})

	ticker.tick()
	queue.awaitCompletion(time.Second)
}
