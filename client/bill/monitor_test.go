// +build !noclientbilltest

package bill

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

type pwStore struct{}

func (s *pwStore) Get() string { return "test-password" }

var (
	conf struct {
		ClientBilling *Config
		DB            *data.DBConfig
		Job           *job.Config
		Log           *log.WriterConfig
		Proc          *proc.Config
		TestTimeout   uint64
	}

	logger log.Logger
	db     *reform.DB
	pr     *proc.Processor
	queue  job.Queue
	pws    *pwStore

	autoincreaseEnabled = &data.Setting{
		Key:         data.SettingClientAutoincreaseDeposit,
		Value:       "false",
		Permissions: data.ReadWrite,
		Name:        data.SettingClientAutoincreaseDeposit,
	}
	autoincreaseAt = &data.Setting{
		Key:         data.SettingClientAutoincreaseDepositPercent,
		Value:       "60",
		Permissions: data.ReadWrite,
		Name:        data.SettingClientAutoincreaseDepositPercent,
	}
	defaultGasPrice = &data.Setting{
		Key:         data.SettingDefaultGasPrice,
		Value:       "10",
		Permissions: data.ReadWrite,
		Name:        "default gas price",
	}
)

func newTestMonitor(
	processErrChan, postErrChan chan error) (*Monitor, chan error) {
	mon := NewMonitor(conf.ClientBilling, logger, db, pr, queue,
		"test-psc-address", pws)
	mon.processErrors = processErrChan
	mon.postChequeErrors = postErrChan

	ch := make(chan error)
	go func() { ch <- mon.Run() }()

	return mon, ch
}

func closeTestMonitor(t *testing.T, mon *Monitor, ch chan error) {
	mon.Close()
	util.TestExpectResult(t, "Run", ErrMonitorClosed, <-ch)
}

func newFixture(t *testing.T, db *reform.DB) *data.TestFixture {
	fxt := data.NewTestFixture(t, db)
	fxt.Channel.ServiceStatus = data.ServiceActive
	fxt.Channel.Client = fxt.Channel.Agent
	data.SaveToTestDB(t, db, fxt.Channel)
	return fxt
}

func newWaitGroup() *sync.WaitGroup {
	wg := sync.WaitGroup{}
	wg.Add(1)
	return &wg
}

func awaitingPosting(wg *sync.WaitGroup, postErrors chan error) {
	defer wg.Done()

	select {
	case <-postErrors:
	case <-time.After(
		time.Duration(conf.TestTimeout) * time.Second):
	}
}

func awaitingGoodPosting(wg *sync.WaitGroup, postErrors chan error) {
	defer wg.Done()
	for {
		select {
		case e := <-postErrors:
			if e == nil {
				return
			}
		case <-time.After(
			time.Duration(conf.TestTimeout) * time.Second):
			return
		}
	}
}

func TestTerminateInactiveChannel(t *testing.T) {
	fxt := newFixture(t, db)
	defer fxt.Close()
	// Insert settings for proper work of monitor
	data.InsertToTestDB(t, fxt.DB, autoincreaseEnabled, autoincreaseAt)
	defer data.DeleteFromTestDB(t, fxt.DB, autoincreaseEnabled, autoincreaseAt)

	fxt.Channel.TotalDeposit = 10
	oneSec := uint64(1)
	fxt.Offering.MaxInactiveTimeSec = oneSec
	session := data.NewTestSession(fxt.Channel.ID)
	// Fake session stopped a day ago.
	stopped := time.Now().AddDate(0, 0, -1)
	session.Stopped = &stopped
	data.SaveToTestDB(t, db, fxt.Channel, fxt.Offering, session)
	defer data.DeleteFromTestDB(t, db, session)
	time.Sleep(time.Second)
	runMonitorAndExpectJobs(t, fxt.Channel.ID, data.JobClientPreServiceTerminate)
}

func TestTerminateCompletedChannel(t *testing.T) {
	fxt := newFixture(t, db)
	defer fxt.Close()
	// Insert settings for proper work of monitor
	data.InsertToTestDB(t, fxt.DB, autoincreaseEnabled, autoincreaseAt)
	defer data.DeleteFromTestDB(t, fxt.DB, autoincreaseEnabled, autoincreaseAt)

	fxt.Channel.TotalDeposit = 10
	fxt.Channel.ReceiptBalance = 10
	data.SaveToTestDB(t, db, fxt.Channel)

	runMonitorAndExpectJobs(t, fxt.Channel.ID, data.JobClientPreServiceTerminate)
}

func TestAutoIncreaseDeposit(t *testing.T) {
	fxt := newFixture(t, db)
	defer fxt.Close()

	autoincrease := *autoincreaseEnabled
	autoincrease.Value = "true"
	fxt.Channel.TotalDeposit = fxt.Offering.MinUnits*fxt.Offering.UnitPrice + fxt.Offering.SetupPrice
	fxt.Channel.ReceiptBalance = uint64(float64(fxt.Channel.TotalDeposit) * 0.7)
	s := data.NewTestSession(fxt.Channel.ID)
	s.LastUsageTime = time.Now()
	// 70% used, need to auto increase
	s.UnitsUsed = uint64(float64(fxt.Offering.MinUnits) * 0.7)
	data.InsertToTestDB(t, fxt.DB, &autoincrease, autoincreaseAt, defaultGasPrice, s)
	defer data.DeleteFromTestDB(t, fxt.DB, &autoincrease, autoincreaseAt, defaultGasPrice, s)
	data.SaveToTestDB(t, fxt.DB, fxt.Channel)

	job := runMonitorAndExpectJobs(t, fxt.Channel.ID, data.JobClientPreChannelTopUp)
	var jdata data.JobTopUpChannelData
	json.Unmarshal(job.Data, &jdata)
	// Increase deposit twice, eg deposit equals to channels current total deposit.
	if got, exp := jdata.Deposit, fxt.Channel.TotalDeposit; got != exp {
		t.Fatalf("deposit=%d, want %d", got, exp)
	}
}

func runMonitorAndExpectJobs(t *testing.T, channelID, jType string) *data.Job {
	processSig := make(chan error)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		select {
		case <-processSig:
		case <-time.After(
			time.Duration(conf.TestTimeout) * time.Second):
		}
	}()

	mon, ch := newTestMonitor(processSig, nil)
	defer closeTestMonitor(t, mon, ch)

	wg.Wait()

	jobs, err := db.FindAllFrom(data.JobTable, "related_id", channelID)

	util.TestExpectResult(t, "Find jobs for channel", nil, err)

	var recs []reform.Record
	for _, v := range jobs {
		recs = append(recs, v.(reform.Record))
	}
	defer data.DeleteFromTestDB(t, db, recs...)

	var ret *data.Job
	for _, v := range jobs {
		if v.(*data.Job).Type == jType {
			ret = v.(*data.Job)
		}
	}

	if ret == nil {
		t.Fatalf("job %s not created", jType)
	}
	return ret
}

func expectBalance(t *testing.T, fxt *data.TestFixture, expected uint64) {
	data.ReloadFromTestDB(t, db, fxt.Channel)
	if fxt.Channel.ReceiptBalance != expected {
		t.Fatalf("unexpected receipt balance: %d (expected %d)",
			fxt.Channel.ReceiptBalance, expected)
	}
}

func TestPayment(t *testing.T) {
	fxt := newFixture(t, db)
	defer fxt.Close()
	// Insert settings for proper work of monitor
	data.InsertToTestDB(t, fxt.DB, autoincreaseEnabled, autoincreaseAt)
	defer data.DeleteFromTestDB(t, fxt.DB, autoincreaseEnabled, autoincreaseAt)

	fxt.Offering.UnitPrice = 1
	fxt.Offering.SetupPrice = 2
	fxt.Offering.BillingInterval = 2
	fxt.Offering.MaxInactiveTimeSec = 1000

	fxt.Channel.TotalDeposit = 10
	fxt.Channel.ReceiptBalance = 4

	sess := data.NewTestSession(fxt.Channel.ID)
	sess.UnitsUsed = 4
	sess.LastUsageTime = time.Now()

	data.SaveToTestDB(t, db, fxt.Offering, fxt.Channel, sess)
	defer data.DeleteFromTestDB(t, db, sess)

	processErrors := make(chan error)
	postErrors := make(chan error)

	mon, ch := newTestMonitor(processErrors, postErrors)
	defer closeTestMonitor(t, mon, ch)

	mtx := sync.Mutex{}
	called := false
	err := fmt.Errorf("some error")
	mon.post = func(db *reform.DB, channel string, pscAddr data.HexString,
		pass string, amount uint64, tls bool, timeout uint,
		pr *proc.Processor) error {
		mtx.Lock()
		defer mtx.Unlock()
		called = true
		return err
	}

	wg := newWaitGroup()
	go awaitingPosting(wg, postErrors)

	sess2 := data.NewTestSession(fxt.Channel.ID)
	sess2.UnitsUsed = 2
	data.SaveToTestDB(t, db, sess2)
	defer data.DeleteFromTestDB(t, db, sess2)

	wg.Wait()

	mtx.Lock()
	if !called {
		t.Fatalf("no payment triggered")
	}
	mtx.Unlock()
	expectBalance(t, fxt, 4)

	wg = newWaitGroup()
	go awaitingGoodPosting(wg, postErrors)

	mtx.Lock()
	err = nil
	mtx.Unlock()

	wg.Wait()

	expectBalance(t, fxt, 8)
}

func TestMain(m *testing.M) {
	conf.ClientBilling = NewConfig()
	conf.Log = log.NewWriterConfig()
	conf.DB = data.NewDBConfig()
	conf.Proc = proc.NewConfig()
	args := &util.TestArgs{
		Conf: &conf,
	}
	util.ReadTestArgs(args)

	var err error
	logger, err = log.NewTestLogger(conf.Log, args.Verbose)
	if err != nil {
		panic(err)
	}

	db = data.NewTestDB(conf.DB)
	queue = job.NewQueue(conf.Job, logger, db, nil)
	pr = proc.NewProcessor(conf.Proc, db, queue)
	pws = &pwStore{}

	os.Exit(m.Run())
}
