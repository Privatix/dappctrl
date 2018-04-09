// +build !nojobtest

package job

import (
	"errors"
	"math/rand"
	"os"
	"testing"
	"time"

	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

type jobTestConfig struct {
	StressJobs uint
}

func newJobTestConfig() *jobTestConfig {
	return &jobTestConfig{
		StressJobs: 100,
	}
}

var (
	conf struct {
		DB      *data.DBConfig
		Job     *Config
		JobTest *jobTestConfig
		Log     *util.LogConfig
	}

	logger *util.Logger
	db     *reform.DB
)

func add(t *testing.T, queue *Queue, job *data.Job, expected error) {
	if err := queue.Add(job); err != expected {
		if err == nil {
			queue.db.Delete(job)
		}
		util.ExpectResult(t, "Add", expected, err)
	}
}

func createJob() *data.Job {
	return &data.Job{
		Type:        data.JobClientPreChannelCreate,
		RelatedID:   util.NewUUID(),
		RelatedType: data.JobChannel,
		CreatedBy:   data.JobUser,
		CreatedAt:   time.Now(),
	}
}

func TestAdd(t *testing.T) {
	queue := NewQueue(conf.Job, logger, db, nil)
	defer queue.Close()

	job := createJob()
	add(t, queue, job, nil)
	defer db.Delete(job)

	rid := job.RelatedID
	job = createJob()
	job.RelatedID = rid
	add(t, queue, job, ErrDuplicatedJob)
	defer db.Delete(job)

	job = createJob()
	job.Type = data.JobClientAfterChannelCreate
	add(t, queue, job, nil)
	defer db.Delete(job)
}

func TestHandlerNotFound(t *testing.T) {
	queue := NewQueue(conf.Job, logger, db, nil)

	job := createJob()
	add(t, queue, job, nil)
	defer db.Delete(job)

	util.ExpectResult(t, "Process", ErrHandlerNotFound, queue.Process())
}

func waitForJob(queue *Queue, job *data.Job, ch chan<- error) {
	for {
		if err := db.FindByPrimaryKeyTo(job, job.ID); err != nil {
			queue.Close()
			ch <- err
			return
		}

		if job.Status != data.JobActive {
			queue.Close()
			ch <- nil
			return
		}

		time.Sleep(time.Millisecond)
	}
}

func TestFailure(t *testing.T) {
	makeHandler := func(limit uint8) Handler {
		return func(j *data.Job) error {
			if j.TryCount+1 < limit {
				return errors.New("some error")
			}
			return nil
		}
	}

	handlerMap := HandlerMap{
		data.JobClientPreChannelCreate: makeHandler(conf.Job.TryLimit),
	}
	queue := NewQueue(conf.Job, logger, db, handlerMap)

	job := createJob()
	add(t, queue, job, nil)
	defer db.Delete(job)

	ch := make(chan error)
	go waitForJob(queue, job, ch)
	util.ExpectResult(t, "Process", ErrQueueClosed, queue.Process())
	util.ExpectResult(t, "waitForJob", nil, <-ch)
	if job.Status != data.JobDone {
		t.Fatalf("job status is not done: %s", job.Status)
	}

	job.TryCount = 0
	job.Status = data.JobActive
	handlerMap[data.JobClientPreChannelCreate] =
		makeHandler(conf.Job.TryLimit + 1)
	util.ExpectResult(t, "Save", nil, db.Save(job))

	go waitForJob(queue, job, ch)
	util.ExpectResult(t, "Process", ErrQueueClosed, queue.Process())
	util.ExpectResult(t, "waitForJob", nil, <-ch)
	if job.Status != data.JobFailed {
		t.Fatalf("job status is not failed: %s", job.Status)
	}
}

func TestStress(t *testing.T) {
	started := time.Now()
	numStressJobs := int(conf.JobTest.StressJobs)

	ch := make(chan struct{})
	handler := func(j *data.Job) error {
		if rand.Uint32()%1 == 0 {
			time.Sleep(time.Millisecond)
		}

		if j.TryCount+1 < conf.Job.TryLimit && rand.Uint32()%2 == 0 {
			return errors.New("some error")
		}

		// Ignore stale jobs not to deadlock the test.
		if j.CreatedAt.After(started) {
			ch <- struct{}{}
		}

		return nil
	}

	queue := NewQueue(conf.Job, logger, db,
		HandlerMap{data.JobClientPreChannelCreate: handler})

	ch2 := make(chan error)
	go func() {
		ch2 <- queue.Process()
	}()

	for i := 0; i < numStressJobs; i++ {
		job := createJob()
		add(t, queue, job, nil)
		defer db.Delete(job)
	}

	for i := 0; i < numStressJobs; i++ {
		<-ch
	}

	queue.Close()
	util.ExpectResult(t, "Process", ErrQueueClosed, <-ch2)
}

func TestMain(m *testing.M) {
	conf.Job = NewConfig()
	conf.Log = util.NewLogConfig()
	conf.JobTest = newJobTestConfig()
	conf.DB = data.NewDBConfig()
	util.ReadTestConfig(&conf)

	logger = util.NewTestLogger(conf.Log)
	db = data.NewTestDB(conf.DB, logger)

	os.Exit(m.Run())
}
