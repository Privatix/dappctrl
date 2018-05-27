// +build !noclientbilltest

package bill

import (
	"fmt"
	"os"
	"testing"
	"time"

	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util"
)

type pwStore struct{}

func (s *pwStore) Get() string { return "test-password" }

type testConfig struct {
	MonitorStartupDelay uint // In milliseconds.
	ReactionDelay       uint // In milliseconds.
}

func newTestConfig() *testConfig {
	return &testConfig{
		MonitorStartupDelay: 10,
		ReactionDelay:       2,
	}
}

var (
	conf struct {
		ClientBilling     *Config
		ClientBillingTest *testConfig
		DB                *data.DBConfig
		Job               *job.Config
		Log               *util.LogConfig
		Proc              *proc.Config
	}

	logger *util.Logger
	db     *reform.DB
	pr     *proc.Processor
	pws    *pwStore
)

func newTestMonitor() (*Monitor, chan error) {
	mon := NewMonitor(conf.ClientBilling,
		logger, db, pr, "test-psc-address", pws)

	ch := make(chan error)
	go func() { ch <- mon.Run() }()

	time.Sleep(time.Duration(conf.ClientBillingTest.MonitorStartupDelay) *
		time.Millisecond)

	return mon, ch
}

func closeTestMonitor(t *testing.T, mon *Monitor, ch chan error) {
	mon.Close()
	util.TestExpectResult(t, "Run", ErrMonitorClosed, <-ch)
}

func delay() {
	time.Sleep(time.Duration(conf.ClientBillingTest.ReactionDelay) *
		time.Millisecond)
}

func newFixture(t *testing.T, db *reform.DB) *data.TestFixture {
	fxt := data.NewTestFixture(t, db)
	fxt.Channel.ServiceStatus = data.ServiceActive
	fxt.Channel.Client = fxt.Channel.Agent
	data.SaveToTestDB(t, db, fxt.Channel)
	return fxt
}

func TestTerminate(t *testing.T) {
	fxt := newFixture(t, db)
	defer fxt.Close()

	fxt.Channel.TotalDeposit = 10
	fxt.Channel.ReceiptBalance = 10
	data.SaveToTestDB(t, db, fxt.Channel)

	mon, ch := newTestMonitor()
	defer closeTestMonitor(t, mon, ch)

	delay()

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
		t.Fatalf("unexpected receipt balance: %d", expected)
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

	mon, ch := newTestMonitor()
	defer closeTestMonitor(t, mon, ch)

	called := false
	err := fmt.Errorf("some error")
	mon.post = func(db *reform.DB, channel, pscAddr, pass string,
		amount uint64, tls bool, timeout uint) error {
		called = true
		return err
	}

	delay()
	if called {
		t.Fatalf("unexpected payment triggering")
	}

	sess2 := data.NewTestSession(fxt.Channel.ID)
	sess2.UnitsUsed = 2
	data.SaveToTestDB(t, db, sess2)
	defer data.DeleteFromTestDB(t, db, sess2)

	delay()
	if !called {
		t.Fatalf("no payment triggered")
	}
	expectBalance(t, fxt, 4)

	err = nil
	delay()
	expectBalance(t, fxt, 8)
}

func TestMain(m *testing.M) {
	conf.ClientBilling = NewConfig()
	conf.ClientBillingTest = newTestConfig()
	conf.Log = util.NewLogConfig()
	conf.DB = data.NewDBConfig()
	conf.Proc = proc.NewConfig()
	util.ReadTestConfig(&conf)

	logger = util.NewTestLogger(conf.Log)
	db = data.NewTestDB(conf.DB, logger)
	queue := job.NewQueue(conf.Job, logger, db, nil)
	pr = proc.NewProcessor(conf.Proc, queue)
	pws = &pwStore{}

	os.Exit(m.Run())
}
