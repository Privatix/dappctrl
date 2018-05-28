// +build !noagentbilltest

package billing

import (
	"os"
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util"
)

var (
	conf struct {
		BillingTest *billingTestConfig
		DB          *data.DBConfig
		Job         *job.Config
		Log         *util.LogConfig
		Pc          *proc.Config
	}

	db     *reform.DB
	logger *util.Logger
	mon    *Monitor
	pr     *proc.Processor
)

const (
	jobRelatedID = "related_id"

	testPassword = "test-password"
)

const (
	errAccNotUsed = "Billing processes channels with" +
		" not used account"

	errChStatusPending = "Billing processes channels with channel" +
		" status is pending"

	errDB = "Failed to read channel information" +
		" from the database"

	errTestResult = "Wrong result"
)

type billingTestConfig struct {
	Offer   offer
	Session session
	Channel channel
}

type offer struct {
	MaxUnit            uint64
	MaxInactiveTimeSec uint64
	UnitPrice          uint64
	BigLag             uint
	SmallLag           uint
}

type session struct {
	UnitsUsed       uint64
	SecondsConsumed uint64
}

type channel struct {
	SmallDeposit uint64
	MidDeposit   uint64
	BigDeposit   uint64
}

type testFixture struct {
	t        *testing.T
	client   *data.User
	agent    *data.Account
	product  *data.Product
	template *data.Template
	testObjs []reform.Record
	chs      []*data.Channel
}

func newTestMonitor(interval time.Duration, db *reform.DB,
	logger *util.Logger, pc *proc.Processor) *Monitor {
	mon, err := NewMonitor(interval, db, logger, pc)
	if err != nil {
		panic(err)
	}
	return mon
}

func newBillingTestConfig() *billingTestConfig {
	return &billingTestConfig{}
}

func newFixture(t *testing.T) *testFixture {
	clientAcc := data.NewTestAccount(testPassword)

	client := data.NewTestUser()

	client.PublicKey = clientAcc.PublicKey

	client.EthAddr = clientAcc.EthAddr

	agent := data.NewTestAccount(testPassword)

	product := data.NewTestProduct()

	template := data.NewTestTemplate(data.TemplateOffer)

	data.InsertToTestDB(t, db, client, agent, product, template)

	return &testFixture{
		t:        t,
		client:   client,
		agent:    agent,
		product:  product,
		template: template,
	}
}

func (f *testFixture) addTestObjects(testObjs []reform.Record) {
	data.SaveToTestDB(f.t, db, testObjs...)

	f.testObjs = append(f.testObjs, testObjs...)
}

func (f *testFixture) clean() {
	records := append([]reform.Record{}, f.client, f.agent,
		f.product, f.template)

	records = append(records, f.testObjs...)

	reverse(records)

	for _, v := range records {
		if err := db.Delete(v); err != nil {
			f.t.Fatalf("failed to delete %T: %s", v, err)
		}
	}
}

func (f *testFixture) checkJob(t *testing.T, ch int,
	checker func(t *testing.T), status string) {
	checker(t)

	if !done(f.chs[ch].ID, status) {
		t.Fatal(errTestResult)
	}
}

func (f *testFixture) checkChanStatus(t *testing.T, ch int,
	checker func(t *testing.T), status string) {
	chStatus(t, f.chs[ch], data.ChannelPending)

	checker(t)

	if done(f.chs[ch].ID, status) {
		t.Fatal(errChStatusPending)
	}

	chStatus(t, f.chs[ch], data.ChannelActive)
}

func (f *testFixture) checkAcc(t *testing.T, ch int,
	checker func(t *testing.T), status string) {
	accNotUse(t, f.agent)

	checker(t)

	if done(f.chs[ch].ID, status) {
		t.Fatal(errAccNotUsed)
	}
}

func reverse(rs []reform.Record) {
	last := len(rs) - 1

	for i := 0; i < len(rs)/2; i++ {
		rs[i], rs[last-i] = rs[last-i], rs[i]
	}
}

func sesFabric(chanID string, secondsConsumed,
	unitsUsed uint64, adjustTime int64, num int) (
	sessions []*data.Session) {
	if num <= 0 {
		return sessions
	}

	for i := 0; i <= num; i++ {
		curTime := time.Now()

		if adjustTime != 0 {
			curTime = curTime.Add(
				time.Second * time.Duration(adjustTime))
		}

		sessions = append(sessions, &data.Session{
			ID:              util.NewUUID(),
			Channel:         chanID,
			Started:         time.Now(),
			LastUsageTime:   curTime,
			SecondsConsumed: secondsConsumed,
			UnitsUsed:       unitsUsed,
		})
	}

	return sessions
}

func done(id, status string) bool {
	var j data.Job

	if err := db.FindOneTo(&j, jobRelatedID, id); err != nil {
		return false
	}

	defer remJob(j)

	if j.CreatedBy != data.JobBillingChecker ||
		j.Type != status ||
		j.Status != data.JobActive {
		return false
	}

	return true
}

func remJob(j data.Job) {
	db.Delete(&j)
}

func chStatus(t *testing.T, ch *data.Channel, status string) {
	ch.ChannelStatus = status
	if err := db.Update(ch); err != nil {
		t.Fatal(err)
	}
}

func accNotUse(t *testing.T, acc *data.Account) {
	acc.InUse = false
	if err := db.Update(acc); err != nil {
		t.Fatal(err)
	}
}

func TestNewMonitor(t *testing.T) {
	_, err := NewMonitor(-1, db, logger, pr)
	if err == nil || err != ErrInput {
		t.Fatal(errTestResult)
	}

	_, err = NewMonitor(time.Second, nil, logger, pr)
	if err == nil || err != ErrInput {
		t.Fatal(errTestResult)
	}

	_, err = NewMonitor(time.Second, db, nil, pr)
	if err == nil || err != ErrInput {
		t.Fatal(errTestResult)
	}

	_, err = NewMonitor(time.Second, db, logger, nil)
	if err == nil || err != ErrInput {
		t.Fatal(errTestResult)
	}
}

func TestMonitorRun(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal(errTestResult)
		}
	}()

	mon, err := NewMonitor(time.Second, &reform.DB{}, logger, pr)
	if err != nil {
		panic(err)
	}

	if err := mon.Run(); err == nil {
		t.Fatal(errTestResult)
	}
}

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.Job = job.NewConfig()
	conf.Log = util.NewLogConfig()
	conf.Pc = proc.NewConfig()
	conf.BillingTest = newBillingTestConfig()

	util.ReadTestConfig(&conf)

	logger = util.NewTestLogger(conf.Log)

	db = data.NewTestDB(conf.DB, logger)
	defer data.CloseDB(db)

	queue := job.NewQueue(conf.Job, logger, db, nil)
	pr = proc.NewProcessor(conf.Pc, queue)

	mon = newTestMonitor(time.Second, db, logger, pr)

	os.Exit(m.Run())
}
