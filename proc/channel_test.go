// +build !noproctest

package proc

import (
	"os"
	"testing"
	"time"

	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util"
)

var (
	conf struct {
		DB   *data.DBConfig
		Job  *job.Config
		Log  *util.LogConfig
		Proc *Config
	}

	db   *reform.DB
	proc *Processor
)

func newTestJob(channel string) *data.Job {
	return &data.Job{
		ID:          util.NewUUID(),
		Status:      data.JobActive,
		RelatedType: data.JobChannel,
		RelatedID:   channel,
		CreatedBy:   data.JobUser,
		Data:        []byte("{}"),
	}
}

type channelActionFunc func(id, jobCreator string) (string, error)

func testChannelAction(t *testing.T, channelAction channelActionFunc,
	funcName, jobType, badServiceStatus, goodServiceStatus,
	jobTypeToCheck string, cancel bool) {
	_, err := channelAction(util.NewUUID(), data.JobUser)
	util.TestExpectResult(t, funcName, reform.ErrNoRows, err)

	fxt := data.NewTestFixture(t, db)
	defer fxt.Close()

	fxt.Channel.ServiceStatus = badServiceStatus
	data.SaveToTestDB(t, db, fxt.Channel)
	_, err = channelAction(fxt.Channel.ID, data.JobUser)
	util.TestExpectResult(t, funcName, ErrBadServiceStatus, err)

	job := newTestJob(fxt.Channel.ID)

	expected := ErrActiveJobsExist
	if len(jobTypeToCheck) != 0 {
		job.Type = jobTypeToCheck
		expected = ErrSameJobExists
	}

	fxt.Channel.ServiceStatus = goodServiceStatus
	data.SaveToTestDB(t, db, fxt.Channel, job)
	defer data.DeleteFromTestDB(t, db, job)

	_, err = channelAction(fxt.Channel.ID, data.JobUser)
	util.TestExpectResult(t, funcName, expected, err)

	if len(jobTypeToCheck) != 0 {
		job.Type = data.JobAgentPreServiceSuspend
	} else {
		job.Status = data.JobDone
	}
	data.SaveToTestDB(t, db, job)

	before := time.Now()
	id, err := channelAction(fxt.Channel.ID, data.JobBCMonitor)
	after := time.Now()
	util.TestExpectResult(t, funcName, nil, err)

	if len(jobTypeToCheck) != 0 {
		data.ReloadFromTestDB(t, db, job)
		if job.Status != data.JobCanceled {
			t.Fatalf("job wasn't cancelled")
		}
	}

	job = &data.Job{ID: id}
	defer data.DeleteFromTestDB(t, db, job)

	data.ReloadFromTestDB(t, db, job)
	if job.Type != jobType || job.RelatedID != fxt.Channel.ID ||
		job.RelatedType != data.JobChannel ||
		job.CreatedAt.Before(before) || job.CreatedAt.After(after) ||
		job.CreatedBy != data.JobBCMonitor {
		t.Fatalf("bad job data")
	}
}

func TestSuspendChannel(t *testing.T) {
	testChannelAction(t, proc.SuspendChannel, "SuspendChannel",
		data.JobAgentPreServiceSuspend, data.ServiceSuspended,
		data.ServiceActive, "", false)
}

func TestActivateChannel(t *testing.T) {
	testChannelAction(t, proc.ActivateChannel, "ActivateChannel",
		data.JobAgentPreServiceUnsuspend, data.ServiceActive,
		data.ServiceSuspended, "", false)
}

func TestTerminateChannel(t *testing.T) {
	testChannelAction(t, proc.TerminateChannel, "TerminateChannel",
		data.JobAgentPreServiceTerminate, data.ServiceTerminated,
		data.ServiceSuspended, data.JobAgentPreServiceTerminate, true)
}

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.Job = job.NewConfig()
	conf.Log = util.NewLogConfig()
	conf.Proc = NewConfig()
	util.ReadTestConfig(&conf)

	logger := util.NewTestLogger(conf.Log)

	db = data.NewTestDB(conf.DB, logger)
	defer data.CloseDB(db)

	queue := job.NewQueue(conf.Job, logger, db, nil)
	proc = NewProcessor(conf.Proc, queue)

	os.Exit(m.Run())
}
