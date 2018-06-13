package job

import (
	"errors"
	"hash/crc32"
	"runtime"
	"sync"
	"time"

	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

// Errors.
var (
	ErrAlreadyProcessing = errors.New("already processing")
	ErrDuplicatedJob     = errors.New("duplicated job")
	ErrHandlerNotFound   = errors.New("job handler not found")
	ErrQueueClosed       = errors.New("queue closed")
)

// Handler is a job handler function.
type Handler func(j *data.Job) error

// HandlerMap is a map of job handlers.
type HandlerMap map[string]Handler

// TypeConfig is a configuration for specific job type.
type TypeConfig struct {
	TryLimit   uint8 // Default number of tries to complete job.
	TryPeriod  uint  // Default retry period, in milliseconds.
	Duplicated bool  // Whether do or do not check for duplicates.
}

// Config is a job queue configuration.
type Config struct {
	CollectJobs   uint // Number of jobs to process for collect-iteration.
	CollectPeriod uint // Collect-iteration period, in milliseconds.
	WorkerBufLen  uint // Worker buffer length.
	Workers       uint // Number of workers, 0 means number of CPUs.

	TypeConfig                       // Default type configuration.
	Types      map[string]TypeConfig // Type-specific overrides.
}

// NewConfig creates a default job queue configuration.
func NewConfig() *Config {
	return &Config{
		CollectJobs:   100,
		CollectPeriod: 1000,
		WorkerBufLen:  10,
		Workers:       0,

		TypeConfig: TypeConfig{
			TryLimit:  3,
			TryPeriod: 60000,
		},
		Types: make(map[string]TypeConfig),
	}
}

type workerIO struct {
	job    chan string
	result chan error
}

// Queue is a job processing queue.
type Queue struct {
	conf     *Config
	logger   *util.Logger
	db       *reform.DB
	handlers HandlerMap
	mtx      sync.Mutex // Prevents races when starting and stopping.
	exit     chan struct{}
	exited   chan struct{}
	workers  []workerIO
}

// NewQueue creates a new job queue.
func NewQueue(conf *Config, logger *util.Logger, db *reform.DB,
	handlers HandlerMap) *Queue {
	return &Queue{
		conf:     conf,
		logger:   logger,
		db:       db,
		handlers: handlers,
	}
}

// Logger is an associated util.Logger instance.
func (q *Queue) Logger() *util.Logger {
	return q.logger
}

// DB is an associated reform.DB instance.
func (q *Queue) DB() *reform.DB {
	return q.db
}

func (q *Queue) checkDuplicated(j *data.Job) error {
	_, err := q.db.SelectOneFrom(data.JobTable,
		"WHERE related_id = $1 AND type = $2", j.RelatedID, j.Type)

	if err == nil {
		return ErrDuplicatedJob
	}

	if err != reform.ErrNoRows {
		return err
	}

	return nil
}

// Add adds a new job to the job queue.
func (q *Queue) Add(j *data.Job) error {
	if !q.typeConfig(j).Duplicated {
		if err := q.checkDuplicated(j); err != nil {
			return err
		}
	}

	j.ID = util.NewUUID()
	j.Status = data.JobActive
	j.CreatedAt = time.Now()

	return q.db.Insert(j)
}

// Close causes currently running Process() function to exit.
func (q *Queue) Close() {
	q.mtx.Lock()
	defer q.mtx.Unlock()

	if q.exit == nil {
		return
	}

	q.exit <- struct{}{}
	<-q.exited
}

// Process fetches active jobs and processes them in parallel. This function
// does not return until an error occurs or Close() is called.
func (q *Queue) Process() error {
	q.mtx.Lock()

	if q.exit != nil {
		q.mtx.Unlock()
		return ErrAlreadyProcessing
	}

	num := int(q.conf.Workers)
	if num == 0 {
		num = runtime.NumCPU()
	}

	// Make sure all workers can signal about errors simultaneously.
	q.exit = make(chan struct{}, num)
	q.exited = make(chan struct{}, 1)

	q.mtx.Unlock()

	q.workers = nil
	for i := 0; i < num; i++ {
		w := workerIO{
			make(chan string, q.conf.WorkerBufLen),
			make(chan error, 1),
		}
		q.workers = append(q.workers, w)
		go q.processWorker(w)
	}

	err := q.processMain()

	// Stop the worker routines.

	for _, w := range q.workers {
		close(w.job)
	}

	for _, w := range q.workers {
		werr := <-w.result
		if werr != nil && err == ErrQueueClosed {
			err = werr
		}
	}

	q.exited <- struct{}{}

	q.mtx.Lock()
	q.exit = nil
	q.mtx.Unlock()

	return err
}

func (q *Queue) uuidWorker(uuid string) workerIO {
	i := int(crc32.ChecksumIEEE([]byte(uuid))) % len(q.workers)
	return q.workers[i]
}

func (q *Queue) checkExit() bool {
	select {
	case <-q.exit:
		return true
	default:
		return false
	}
}

func (q *Queue) processMain() error {
	period := time.Duration(q.conf.CollectPeriod) * time.Millisecond

	for {
		if q.checkExit() {
			return ErrQueueClosed
		}

		started := time.Now()

		rows, err := q.db.Query(`
			SELECT id, related_id FROM jobs
			 WHERE status = $1 AND not_before <= $2
			 ORDER BY related_id, created_at
			 LIMIT $3`, data.JobActive, started, q.conf.CollectJobs)
		if err != nil {
			return err
		}

		for rows.Next() {
			if q.checkExit() {
				return ErrQueueClosed
			}

			var job, related string
			if err = rows.Scan(&job, &related); err != nil {
				return err
			}
			q.uuidWorker(related).job <- job
		}
		if err := rows.Err(); err != nil {
			return err
		}

		time.Sleep(period - time.Now().Sub(started))
	}
}

func (q *Queue) processWorker(w workerIO) {
	var err error
	for err == nil {
		id, ok := <-w.job
		if !ok {
			break
		}

		// Job was collected active, but delivered here with some delay,
		// so make sure it's still relevant.
		var job data.Job
		if err = q.db.FindByPrimaryKeyTo(&job, id); err != nil {
			break
		}
		if job.Status != data.JobActive {
			continue
		}

		handler, ok := q.handlers[job.Type]
		if !ok {
			q.logger.Error("job handler for %s not found", job.Type)
			err = ErrHandlerNotFound
			break
		}

		q.processJob(&job, handler)

		// If job was cancelled while running a handler make sure it
		// won't be retried.
		if job.Status == data.JobActive {
			tx, err := q.db.Begin()
			if err != nil {
				break
			}

			var tmp data.Job
			err = tx.SelectOneTo(&tmp,
				"WHERE id = $1 FOR UPDATE", job.ID)
			if err != nil {
				tx.Rollback()
				break
			}

			if tmp.Status == data.JobCanceled {
				job.Status = data.JobCanceled
			}

			if err := tx.Commit(); err != nil {
				break
			}
		}

		err = q.db.Save(&job)
	}

	if err != nil {
		q.exit <- struct{}{}

		// Make sure the main routine is not blocked passing a job.
		select {
		case <-w.job:
		default:
		}
	}

	w.result <- err
}

func (q *Queue) processJob(job *data.Job, handler Handler) {
	tconf := q.typeConfig(job)

	q.logger.Info("processing job %s(%s)", job.ID, job.Type)
	err := handler(job)

	if err == nil {
		job.Status = data.JobDone
		q.logger.Info("job %s(%s) is done", job.ID, job.Type)
		return
	}

	if tconf.TryLimit != 0 {
		job.TryCount++
	}

	if job.TryCount >= tconf.TryLimit && tconf.TryLimit != 0 {
		job.Status = data.JobFailed
		q.logger.Error("job %s(%s) is failed", job.ID, job.Type)
	} else {
		job.NotBefore = time.Now().Add(
			time.Duration(tconf.TryPeriod) * time.Millisecond)
		q.logger.Warn("retry for job %s(%s) scheduled to %s: %s",
			job.ID, job.Type,
			job.NotBefore.Format(time.RFC3339), err)
	}
}

func (q *Queue) typeConfig(job *data.Job) TypeConfig {
	tconf := q.conf.TypeConfig
	if conf, ok := q.conf.Types[job.Type]; ok {
		tconf = conf
	}
	return tconf
}
