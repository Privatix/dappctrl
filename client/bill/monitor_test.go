// +build !noclientbilltest

package bill

import (
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
		FileLog       *log.FileConfig
		Proc          *proc.Config
		TestTimeout   uint64
	}

	logger log.Logger
	db     *reform.DB
	pr     *proc.Processor
	pws    *pwStore
)

func newTestMonitor(
	processErrChan, postErrChan chan error) (*Monitor, chan error) {
	mon := NewMonitor(conf.ClientBilling, logger, db, pr,
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

func awaitingProcessing(wg *sync.WaitGroup, processErrors chan error) {
	defer wg.Done()

	select {
	case <-processErrors:
	case <-time.After(
		time.Duration(conf.TestTimeout) * time.Second):
	}
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

func TestTerminate(t *testing.T) {
	fxt := newFixture(t, db)
	defer fxt.Close()

	fxt.Channel.TotalDeposit = 10
	fxt.Channel.ReceiptBalance = 10
	data.SaveToTestDB(t, db, fxt.Channel)

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

	jobs, err := db.FindAllFrom(data.JobTable, "related_id", fxt.Channel.ID)
	util.TestExpectResult(t, "Find jobs for channel", nil, err)

	var recs []reform.Record
	for _, v := range jobs {
		recs = append(recs, v.(reform.Record))
	}
	defer data.DeleteFromTestDB(t, db, recs...)

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, but found %d", len(jobs))
	}
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

	fxt.Offering.UnitPrice = 1
	fxt.Offering.SetupPrice = 2
	fxt.Offering.BillingInterval = 2

	fxt.Channel.TotalDeposit = 10
	fxt.Channel.ReceiptBalance = 4

	sess := data.NewTestSession(fxt.Channel.ID)
	sess.UnitsUsed = 4

	data.SaveToTestDB(t, db, fxt.Offering, fxt.Channel, sess)
	defer data.DeleteFromTestDB(t, db, sess)

	processErrors := make(chan error)
	postErrors := make(chan error)

	wg := newWaitGroup()
	go awaitingProcessing(wg, processErrors)

	mon, ch := newTestMonitor(processErrors, postErrors)
	defer closeTestMonitor(t, mon, ch)

	called := false
	err := fmt.Errorf("some error")
	mon.post = func(db *reform.DB, channel, pscAddr, pass string,
		amount uint64, tls bool, timeout uint,
		pr *proc.Processor) error {
		called = true
		return err
	}

	wg.Wait()

	if called {
		t.Fatalf("unexpected payment triggering")
	}

	wg = newWaitGroup()
	go awaitingPosting(wg, postErrors)

	sess2 := data.NewTestSession(fxt.Channel.ID)
	sess2.UnitsUsed = 2
	data.SaveToTestDB(t, db, sess2)
	defer data.DeleteFromTestDB(t, db, sess2)

	wg.Wait()

	if !called {
		t.Fatalf("no payment triggered")
	}
	expectBalance(t, fxt, 4)

	wg = newWaitGroup()
	go awaitingGoodPosting(wg, postErrors)

	err = nil

	wg.Wait()

	expectBalance(t, fxt, 8)
}

func TestMain(m *testing.M) {
	conf.ClientBilling = NewConfig()
	conf.FileLog = log.NewFileConfig()
	conf.DB = data.NewDBConfig()
	conf.Proc = proc.NewConfig()
	util.ReadTestConfig(&conf)

	l, err := log.NewStderrLogger(conf.FileLog)
	if err != nil {
		panic(err)
	}

	logger = l

	db = data.NewTestDB(conf.DB)
	queue := job.NewQueue(conf.Job, logger, db, nil)
	pr = proc.NewProcessor(conf.Proc, db, queue)
	pws = &pwStore{}

	os.Exit(m.Run())
}
