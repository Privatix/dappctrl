// +build !noclientsvcruntest

package svcrun

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

var (
	conf struct {
		DB            *data.DBConfig
		StderrLog     *log.WriterConfig
		Log           *util.LogConfig
		Job           *job.Config
		Proc          *proc.Config
		ServiceRunner *Config
		TestTimeout   uint64
	}

	logger log.Logger
	db     *reform.DB
	pr     *proc.Processor
)

func assertNotRunning(t *testing.T, runner *serviceRunner, channel string) {
	running, err := runner.IsRunning(channel)
	util.TestExpectResult(t, "Check if running", nil, err)
	if running {
		t.Fatalf("service is running")
	}
}

func assertJobAdded(t *testing.T, channel string) {
	var job data.Job
	data.FindInTestDB(t, db, &job, "related_id", channel)
	defer data.DeleteFromTestDB(t, db, &job)

	if job.Type != data.JobClientPreServiceSuspend {
		t.Fatalf("unexpected job type: %s", job.Type)
	}
}

func newTestServiceRunner() *serviceRunner {
	runner := NewServiceRunner(
		conf.ServiceRunner, logger, db, pr).(*serviceRunner)

	runner.newCmd = func(
		name string, args []string, channel string) *exec.Cmd {
		return exec.Command(name, args...)
	}

	return runner
}

func newTestFixture(t *testing.T) *data.TestFixture {
	fxt := data.NewTestFixture(t, db)
	fxt.Channel.ServiceStatus = data.ServiceActive
	data.SaveToTestDB(t, db, fxt.Channel)
	return fxt
}

func TestStart(t *testing.T) {
	fxt := newTestFixture(t)
	defer fxt.Close()

	runner := newTestServiceRunner()
	defer runner.StopAll()

	runner.done = make(chan bool)

	err := runner.Start(fxt.Channel.ID)
	util.TestExpectResult(t, "Start service", nil, err)

	err = runner.Start(fxt.Channel.ID)
	util.TestExpectResult(t, "Start service",
		ErrAlreadyStarted, err)

	select {
	case <-runner.done:
	case <-time.After(time.Duration(conf.TestTimeout) * time.Second):
	}

	assertNotRunning(t, runner, fxt.Channel.ID)
	assertJobAdded(t, fxt.Channel.ID)
}

func TestStop(t *testing.T) {
	fxt := newTestFixture(t)
	defer fxt.Close()

	runner := newTestServiceRunner()
	defer runner.StopAll()

	err := runner.Start(fxt.Channel.ID)
	util.TestExpectResult(t, "Start service",
		nil, err)

	util.TestExpectResult(t, "Stop service",
		nil, runner.Stop(fxt.Channel.ID))

	assertNotRunning(t, runner, fxt.Channel.ID)
	assertJobAdded(t, fxt.Channel.ID)
}

func TestStopAll(t *testing.T) {
	fxt := newTestFixture(t)
	defer fxt.Close()

	ch := *fxt.Channel
	ch.ID = util.NewUUID()
	data.InsertToTestDB(t, db, &ch)
	defer data.DeleteFromTestDB(t, db, &ch)

	conf2 := conf.ServiceRunner
	if sconf, ok := conf2.Services[fxt.Offering.ServiceName]; ok {
		sconf.Single = false
		conf2.Services = map[string]ServiceConfig{
			fxt.Offering.ServiceName: sconf,
		}
	} else {
		t.Fatalf("no service config found")
	}

	runner := newTestServiceRunner()

	err := runner.Start(fxt.Channel.ID)
	util.TestExpectResult(t, "Start service 1", nil, err)

	err = runner.Start(ch.ID)
	util.TestExpectResult(t, "Start service 2", nil, err)

	runner.StopAll()

	assertNotRunning(t, runner, fxt.Channel.ID)
	assertJobAdded(t, fxt.Channel.ID)
	assertNotRunning(t, runner, ch.ID)
	assertJobAdded(t, ch.ID)
}

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.StderrLog = log.NewWriterConfig()
	conf.Job = job.NewConfig()
	conf.Proc = proc.NewConfig()
	conf.ServiceRunner = NewConfig()
	util.ReadTestConfig(&conf)

	l, err := log.NewStderrLogger(conf.StderrLog)
	if err != nil {
		panic(err)
	}

	logger = l

	db = data.NewTestDB(conf.DB)
	defer data.CloseDB(db)

	queue := job.NewQueue(conf.Job, l, db, nil)
	pr = proc.NewProcessor(conf.Proc, db, queue)

	os.Exit(m.Run())
}
