// +build !nojobtest

package job

import (
	"errors"
	"math/rand"
	"os"
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

type testConfig struct {
	StressJobs uint
}

func newTestConfig() *testConfig {
	return &testConfig{
		StressJobs: 100,
	}
}

var (
	conf struct {
		DB      *data.DBConfig
		Job     *Config
		JobTest *testConfig
		FileLog *log.FileConfig
	}
	logger log.Logger
	db     *reform.DB
)

func add(t *testing.T, queue *queue, job *data.Job, expected error) {
	if err := queue.Add(nil, job); err != expected {
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
	queue := NewQueue(conf.Job, logger, db, nil).(*queue)
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
	job.RelatedID = rid
	oldConf := queue.conf.Types[job.Type]
	queue.conf.Types[job.Type] = TypeConfig{Duplicated: true}
	add(t, queue, job, nil)
	defer db.Delete(job)
	queue.conf.Types[job.Type] = oldConf

	job = createJob()
	job.Type = data.JobClientAfterChannelCreate
	add(t, queue, job, nil)
	defer db.Delete(job)
}

func TestHandlerNotFound(t *testing.T) {
	queue := NewQueue(conf.Job, logger, db, nil).(*queue)

	job := createJob()
	add(t, queue, job, nil)
	defer db.Delete(job)

	util.TestExpectResult(t, "Process", ErrHandlerNotFound, queue.Process())
}

func waitForJob(queue *queue, job *data.Job, ch chan<- error) {
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
	data.CleanTestTable(t, db, data.JobTable)

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
	queue := NewQueue(conf.Job, logger, db, handlerMap).(*queue)

	job := createJob()
	add(t, queue, job, nil)
	defer db.Delete(job)

	ch := make(chan error)
	go waitForJob(queue, job, ch)
	logger.Info("-1")
	util.TestExpectResult(t, "Process", ErrQueueClosed, queue.Process())
	logger.Info("-2")
	util.TestExpectResult(t, "waitForJob", nil, <-ch)
	if job.Status != data.JobDone {
		t.Fatalf("job status is not done: %s", job.Status)
	}

	job.TryCount = 0
	job.Status = data.JobActive
	handlerMap[data.JobClientPreChannelCreate] =
		makeHandler(conf.Job.TryLimit + 1)
	util.TestExpectResult(t, "Save", nil, db.Save(job))

	go waitForJob(queue, job, ch)
	logger.Info("1")
	util.TestExpectResult(t, "Process", ErrQueueClosed, queue.Process())
	logger.Info("2")
	util.TestExpectResult(t, "waitForJob", nil, <-ch)
	if job.Status != data.JobFailed {
		t.Fatalf("job status is not failed: %s", job.Status)
	}
}

func TestStress(t *testing.T) {
	data.CleanTestTable(t, db, data.JobTable)

	numStressJobs := int(conf.JobTest.StressJobs)

	ch := make(chan struct{})
	handler := func(j *data.Job) error {
		if rand.Uint32()%1 == 0 {
			time.Sleep(time.Millisecond)
		}

		if j.TryCount+1 < conf.Job.TryLimit && rand.Uint32()%2 == 0 {
			return errors.New("some error")
		}

		ch <- struct{}{}

		return nil
	}

	queue := NewQueue(conf.Job, logger, db,
		HandlerMap{data.JobClientPreChannelCreate: handler}).(*queue)

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

func TestSubscribe(t *testing.T) {
	data.CleanTestTable(t, db, data.JobTable)

	job1 := data.Job{
		Type:        "a",
		RelatedType: data.JobChannel,
		RelatedID:   util.NewUUID(),
		CreatedBy:   data.JobTask,
		Data:        []byte("{}"),
	}

	job2 := job1
	job2.Type = "b"

	var q *queue
	handlers := HandlerMap{
		"a": func(j *data.Job) error {
			if j.TryCount == 0 {
				return errors.New("some error")
			}
			return q.Add(nil, &job2)
		},
		"b": func(j *data.Job) error { return nil },
	}
	q = NewQueue(conf.Job, logger, db, handlers).(*queue)
	defer q.Close()

	type subParams struct {
		job    *data.Job
		result error
	}

	subch := make(chan subParams)
	subf := func(job *data.Job, result error) {
		subch <- subParams{job, result}
	}

	go func() { q.Process() }()

	util.TestExpectResult(t, "Add job", nil, q.Add(nil, &job1))
	util.TestExpectResult(t, "Subscribe", nil,
		q.Subscribe([]string{job1.RelatedID}, "1234", subf))
	util.TestExpectResult(t, "Subscribe", ErrSubscriptionExists,
		q.Subscribe([]string{job1.RelatedID}, "1234", subf))

	if params := <-subch; params.job.ID != job1.ID ||
		params.result == nil || params.result.Error() != "some error" {
		t.Errorf("wrong parameters for the first notification")
	}
	if params := <-subch; params.job.ID != job1.ID || params.result != nil {
		t.Errorf("wrong parameters for the second notification")
	}
	if params := <-subch; params.job.ID != job2.ID || params.result != nil {
		t.Errorf("wrong parameters for the third notification")
	}

	util.TestExpectResult(t, "Unsubscribe", nil,
		q.Unsubscribe([]string{job1.RelatedID}, "1234"))
	util.TestExpectResult(t, "Unsubscribe", ErrSubscriptionNotFound,
		q.Unsubscribe([]string{job1.RelatedID}, "1234"))
}

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.Job = NewConfig()
	conf.JobTest = newTestConfig()
	conf.FileLog = log.NewFileConfig()
	util.ReadTestConfig(&conf)

	l, err := log.NewStderrLogger(conf.FileLog)
	if err != nil {
		panic(err)
	}

	logger = l
	db = data.NewTestDB(conf.DB)

	os.Exit(m.Run())
}
