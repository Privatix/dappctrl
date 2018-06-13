package worker

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/messages"
	"github.com/privatix/dappctrl/messages/offer"
	"github.com/privatix/dappctrl/pay"
	"github.com/privatix/dappctrl/somc"
	"github.com/privatix/dappctrl/util"
)

type testConfig struct {
	DB             *data.DBConfig
	JobHanlderTest *struct {
		SOMCTimeout time.Duration // In seconds.
	}
	Gas       *GasConf
	Job       *job.Config
	Log       *util.LogConfig
	PayServer *pay.Config
	SOMC      *somc.Config
	SOMCTest  *somc.TestConfig
	pscAddr   common.Address
}

func newTestConfig() *testConfig {
	return &testConfig{
		DB:       data.NewDBConfig(),
		Job:      job.NewConfig(),
		Log:      util.NewLogConfig(),
		SOMC:     somc.NewConfig(),
		SOMCTest: somc.NewTestConfig(),
		pscAddr:  common.HexToAddress("0x1"),
	}
}

type workerTest struct {
	db       *reform.DB
	ethBack  *testEthBackend
	fakeSOMC *somc.FakeSOMC
	somcConn *somc.Conn
	worker   *Worker
	gasConf  *GasConf
}

var (
	conf   *testConfig
	db     *reform.DB
	logger *util.Logger
)

func newWorkerTest(t *testing.T) *workerTest {

	fakeSOMC := somc.NewFakeSOMC(t, conf.SOMC.URL,
		conf.SOMCTest.ServerStartupDelay)

	somcConn, err := somc.NewConn(conf.SOMC, logger)
	if err != nil {
		t.Fatal(err)
	}

	jobQueue := job.NewQueue(conf.Job, logger, db, nil)

	ethBack := newTestEthBackend(conf.pscAddr)

	pwdStorage := new(data.PWDStorage)
	pwdStorage.Set(data.TestPassword)

	worker, err := NewWorker(db, somcConn, ethBack, conf.Gas,
		conf.pscAddr, conf.PayServer.Addr,
		pwdStorage, data.TestToPrivateKey)
	if err != nil {
		fakeSOMC.Close()
		somcConn.Close()
		panic(err)
	}
	worker.SetQueue(jobQueue)

	return &workerTest{
		db:       db,
		ethBack:  ethBack,
		fakeSOMC: fakeSOMC,
		somcConn: somcConn,
		worker:   worker,
		gasConf:  conf.Gas,
	}
}

func (e *workerTest) close() {
	e.fakeSOMC.Close()
	e.somcConn.Close()
}

func TestMain(m *testing.M) {
	conf = newTestConfig()

	util.ReadTestConfig(conf)

	var err error

	logger, err = util.NewLogger(conf.Log)
	if err != nil {
		panic(err)
	}

	db, err = data.NewDB(conf.DB, logger)
	if err != nil {
		panic(err)
	}
	defer data.CloseDB(db)

	os.Exit(m.Run())
}

func (e *workerTest) insertToTestDB(t *testing.T, recs ...reform.Struct) {
	data.InsertToTestDB(t, e.db, recs...)
}

func (e *workerTest) deleteFromTestDB(t *testing.T, recs ...reform.Record) {
	data.DeleteFromTestDB(t, e.db, recs...)
}

func (e *workerTest) updateInTestDB(t *testing.T, rec reform.Record) {
	data.SaveToTestDB(t, e.db, rec)
}

func (e *workerTest) findTo(t *testing.T, rec reform.Record, id string) {
	err := e.db.FindByPrimaryKeyTo(rec, id)
	if err != nil {
		t.Fatal("failed to find: ", err)
	}
}

func (e *workerTest) deleteJob(t *testing.T, jobType, relType, relID string) {
	job := &data.Job{}
	err := e.db.SelectOneTo(job,
		"WHERE type=$1 AND status=$2 AND related_type=$3"+
			" AND related_id=$4 AND created_by=$5",
		jobType, data.JobActive, relType,
		relID, data.JobTask)
	if err != nil {
		t.Log(err)
		t.Fatalf("%s job expected (%s)", jobType, util.Caller())
	}
	e.deleteFromTestDB(t, job)
}

func (e *workerTest) deleteEthTx(t *testing.T, jobID string) {
	ethTx := &data.EthTx{}
	if err := e.db.FindOneTo(ethTx, "job", jobID); err != nil {
		t.Fatalf("EthTx for job expected, got: %v (%s)", err,
			util.Caller())
	}
	e.deleteFromTestDB(t, ethTx)
}

func runJob(t *testing.T, workerF func(*data.Job) error, job *data.Job) {
	if err := workerF(job); err != nil {
		t.Fatalf("%v (%s)", err, util.Caller())
	}
}

type workerTestFixture struct {
	*data.TestFixture
	job *data.Job
}

func (e *workerTest) newTestFixture(t *testing.T,
	jobType, relType string) *workerTestFixture {
	f := data.NewTestFixture(t, e.db)

	job := data.NewTestJob(jobType, data.JobBCMonitor, relType)
	switch relType {
	case data.JobChannel:
		job.RelatedID = f.Channel.ID
	case data.JobEndpoint:
		job.RelatedID = f.Endpoint.ID
	case data.JobOfferring:
		job.RelatedID = f.Offering.ID
	case data.JobAccount:
		job.RelatedID = f.Account.ID
	}
	e.insertToTestDB(t, job)

	// Clear call stack.
	e.ethBack.callStack = []testEthBackCall{}

	return &workerTestFixture{f, job}
}

func (e *workerTest) setOfferingHash(t *testing.T, fixture *workerTestFixture) {
	msg := offer.OfferingMessage(fixture.Account, fixture.TemplateOffer,
		fixture.Offering)
	msgBytes, _ := json.Marshal(msg)

	agentKey, _ := e.worker.decryptKeyFunc(fixture.Account.PrivateKey,
		data.TestPassword)

	packed, _ := messages.PackWithSignature(msgBytes, agentKey)

	fixture.Offering.RawMsg = data.FromBytes(packed)
	fixture.Offering.Hash = data.FromBytes(ethcrypto.Keccak256(packed))
	e.updateInTestDB(t, fixture.Offering)
}

func (f *workerTestFixture) close() {
	data.DeleteFromTestDB(f.T, f.DB, f.job)
	f.TestFixture.Close()
}

func (f *workerTestFixture) setJobData(t *testing.T, d interface{}) {
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	f.job.Data = b
	data.SaveToTestDB(t, f.DB, f.job)
}

func testCommonErrors(t *testing.T, workerF func(*data.Job) error, job data.Job) {
	for _, f := range []func(*testing.T, func(*data.Job) error, data.Job){
		testWrongType,
		// testWrongRelatedType,
		// testNoRelatedFound,
	} {
		t.Run(funcName(f), func(t *testing.T) {
			f(t, workerF, job)
		})
	}
}

func funcName(f interface{}) string {
	funcName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	funcName = filepath.Ext(funcName)
	return strings.TrimPrefix(funcName, ".")
}

func testWrongRelatedType(t *testing.T, f func(*data.Job) error, job data.Job) {
	job.RelatedType = "wrong-rel-type"
	if f(&job) != ErrInvalidJob {
		t.Fatal("related type not validated")
	}
}

func testWrongType(t *testing.T, f func(*data.Job) error, job data.Job) {
	job.Type = "wrong-type"
	if err := f(&job); err != ErrInvalidJob {
		t.Fatal("type not validated: ", err, ErrInvalidJob)
	}
}

func testNoRelatedFound(t *testing.T, f func(*data.Job) error, job data.Job) {
	job.RelatedID = util.NewUUID()
	if err := f(&job); err != sql.ErrNoRows {
		t.Fatal("no error on related absence, got: ", err)
	}
}
