// +build !nojobtest

package queue

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

func newJobQueueTestConfig() *jobTestConfig {
	return &jobTestConfig{
		StressJobs: 100,
	}
}

var (
	conf struct {
		DB           *data.DBConfig
		JobQueue     *Config
		JobQueueTest *jobTestConfig
		Log          *util.LogConfig
	}

	logger *util.Logger
	db     *reform.DB
)

func add(t *testing.T, queue *Queue, job *data.Job, expected error) {
	if err := queue.Add(job); err != expected {
		if err == nil {
			queue.db.Delete(job)
		}
		util.TestExpectResult(t, "Add", expected, err)
	}
}

func createJob() *data.Job {
	return &data.Job{
		Type:        data.JobClientPreChannelCreate,
		RelatedID:   util.NewUUID(),
		RelatedType: data.JobChannel,
		CreatedBy:   data.JobUser,
		CreatedAt:   time.Now(),
		Data:        []byte("{}"),
	}
}

func TestAdd(t *testing.T) {
	queue := NewQueue(conf.JobQueue, logger, db, nil)
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
	queue := NewQueue(conf.JobQueue, logger, db, nil)

	job := createJob()
	add(t, queue, job, nil)
	defer db.Delete(job)

	util.TestExpectResult(t, "Process", ErrHandlerNotFound, queue.Process())
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
		data.JobClientPreChannelCreate: makeHandler(conf.JobQueue.TryLimit),
	}
	queue := NewQueue(conf.JobQueue, logger, db, handlerMap)

	job := createJob()
	add(t, queue, job, nil)
	defer db.Delete(job)

	ch := make(chan error)
	go waitForJob(queue, job, ch)
	util.TestExpectResult(t, "Process", ErrQueueClosed, queue.Process())
	util.TestExpectResult(t, "waitForJob", nil, <-ch)
	if job.Status != data.JobDone {
		t.Fatalf("job status is not done: %s", job.Status)
	}

	job.TryCount = 0
	job.Status = data.JobActive
	handlerMap[data.JobClientPreChannelCreate] =
		makeHandler(conf.JobQueue.TryLimit + 1)
	util.TestExpectResult(t, "Save", nil, db.Save(job))

	go waitForJob(queue, job, ch)
	util.TestExpectResult(t, "Process", ErrQueueClosed, queue.Process())
	util.TestExpectResult(t, "waitForJob", nil, <-ch)
	if job.Status != data.JobFailed {
		t.Fatalf("job status is not failed: %s", job.Status)
	}
}

func TestStress(t *testing.T) {
	started := time.Now()
	numStressJobs := int(conf.JobQueueTest.StressJobs)

	ch := make(chan struct{})
	handler := func(j *data.Job) error {
		if rand.Uint32()%1 == 0 {
			time.Sleep(time.Millisecond)
		}

		if j.TryCount+1 < conf.JobQueue.TryLimit && rand.Uint32()%2 == 0 {
			return errors.New("some error")
		}

		// Ignore stale jobs not to deadlock the test.
		if j.CreatedAt.After(started) {
			ch <- struct{}{}
		}

		return nil
	}

	queue := NewQueue(conf.JobQueue, logger, db,
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
	util.TestExpectResult(t, "Process", ErrQueueClosed, <-ch2)
}

func TestMain(m *testing.M) {
	conf.JobQueue = NewConfig()
	conf.Log = util.NewLogConfig()
	conf.JobQueueTest = newJobQueueTestConfig()
	conf.DB = data.NewDBConfig()
	util.ReadTestConfig(&conf)

	logger = util.NewTestLogger(conf.Log)
	db = data.NewTestDB(conf.DB, logger)

	os.Exit(m.Run())
}
