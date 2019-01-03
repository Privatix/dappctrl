package job

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"runtime"
	"sync"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

// Handler is a job handler function.
type Handler func(j *data.Job) error

// HandlerMap is a map of job handlers.
type HandlerMap map[string]Handler

// TypeConfig is a configuration for specific job type.
type TypeConfig struct {
	TryLimit        uint8 // Default number of tries to complete job.
	TryPeriod       uint  // Default retry period, in milliseconds.
	Duplicated      bool  // Whether do or do not check for duplicates.
	FirstStartDelay uint  // Default first run delay after job added, in milliseconds.
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

type subEntry struct {
	subID   string
	subFunc SubFunc
}

// Queue is a job processing queue.
type Queue interface {
	Add(tx *reform.TX, j *data.Job) error
	Process() error
	Close()
	Subscribe(relatedIDs []string, subID string, subFunc SubFunc) error
	Unsubscribe(relatedIDs []string, subID string) error
}

type queue struct {
	conf     *Config
	logger   log.Logger
	db       *reform.DB
	handlers HandlerMap
	mtx      sync.Mutex // Prevents races when starting and stopping.
	exit     chan struct{}
	exited   chan struct{}
	workers  []workerIO
	subsMtx  sync.RWMutex
	subs     map[string][]subEntry
}

// NewQueue creates a new job queue.
func NewQueue(conf *Config, logger log.Logger, db *reform.DB,
	handlers HandlerMap) Queue {
	l := logger.Add("type", "job.Queue")
	return &queue{
		conf:     conf,
		logger:   l,
		db:       db,
		handlers: handlers,
		subs:     map[string][]subEntry{},
	}
}

func (q *queue) checkDuplicated(j *data.Job, logger log.Logger) error {
	jdata := &data.JobData{}
	json.Unmarshal(j.Data, jdata)
	var err error
	if jdata.EthLog != nil {
		_, err = q.db.SelectOneFrom(data.JobTable,
			`WHERE related_id = $1
			    AND type = $2
			    AND ((data->'ethereumLog'->'transactionHash') IS NULL
			    OR data->'ethereumLog'->>'transactionHash'=$3)`,
			j.RelatedID, j.Type, jdata.EthLog.TxHash)
	} else {
		_, err = q.db.SelectOneFrom(data.JobTable,
			`WHERE related_id = $1 AND type = $2
			    AND status = 'active'`, j.RelatedID, j.Type)
	}
	if err == nil {
		return ErrDuplicatedJob
	}

	if err != reform.ErrNoRows {
		logger.Error(err.Error())
		return ErrInternal
	}

	return nil
}

// Add adds a new job to the job queue (optionally within a given transaction).
func (q *queue) Add(tx *reform.TX, j *data.Job) error {
	logger := q.logger.Add("method", "Add", "job", j)
	tconf := q.typeConfig(j)
	if !tconf.Duplicated {
		if err := q.checkDuplicated(j, logger); err != nil {
			return err
		}
	}
	if tconf.FirstStartDelay > 0 {
		j.NotBefore = time.Now().Add(
			time.Duration(tconf.FirstStartDelay) * time.Millisecond)
	}

	j.ID = util.NewUUID()
	j.Status = data.JobActive
	j.CreatedAt = time.Now()

	db := q.db.Querier
	if tx != nil {
		db = tx.Querier
	}

	if err := db.Insert(j); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return nil
}

// Close causes currently running Process() function to exit.
func (q *queue) Close() {
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
func (q *queue) Process() error {
	logger := q.logger.Add("method", "Process")
	q.mtx.Lock()

	if q.exit != nil {
		q.mtx.Unlock()
		return ErrAlreadyProcessing
	}

	num := int(q.conf.Workers)
	if num == 0 {
		num = runtime.NumCPU()
	}
	logger = logger.Add("workerNumber", num)

	// Make sure all workers can signal about errors simultaneously.
	q.exit = make(chan struct{}, num)
	q.exited = make(chan struct{}, 1)

	q.mtx.Unlock()

	q.workers = nil
	for i := 0; i < num; i++ {
		w := workerIO{
			job:    make(chan string, q.conf.WorkerBufLen),
			result: make(chan error, 1),
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

func (q *queue) uuidWorker(uuid string) workerIO {
	i := int(crc32.ChecksumIEEE([]byte(uuid))) % len(q.workers)
	return q.workers[i]
}

func (q *queue) checkExit() bool {
	select {
	case <-q.exit:
		return true
	default:
		return false
	}
}

func (q *queue) processMain() error {
	logger := q.logger.Add("method", "processMain")
	period := time.Duration(q.conf.CollectPeriod) * time.Millisecond

	for {
		if q.checkExit() {
			return ErrQueueClosed
		}

		started := time.Now()

		rows, err := q.db.Query(`
			SELECT id, related_id FROM (
			  SELECT DISTINCT ON (related_id) *
			    FROM jobs
			   WHERE status = $1
			   ORDER BY related_id, created_at) AS ordered
			 WHERE not_before <= $2
			 LIMIT $3`, data.JobActive, started, q.conf.CollectJobs)
		if err != nil {
			logger.Error(err.Error())
			return ErrInternal
		}

		for rows.Next() {
			if q.checkExit() {
				logger.Debug(ErrQueueClosed.Error())
				return ErrQueueClosed
			}

			var job, related string
			if err = rows.Scan(&job, &related); err != nil {
				logger.Error(err.Error())
				return ErrInternal
			}
			q.uuidWorker(related).job <- job
		}
		if err := rows.Err(); err != nil {
			logger.Error(err.Error())
			return ErrInternal
		}

		time.Sleep(period - time.Now().Sub(started))
	}
}

func (q *queue) processWorker(w workerIO) {
	logger := q.logger.Add("method", "processWorker")
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
		if job.Status != data.JobActive || time.Now().Before(job.NotBefore) {
			continue
		}

		logger := logger.Add("job", id, "type", job.Type)

		handler, ok := q.handlers[job.Type]
		if !ok {
			logger.Error("job handler not found")
			err = ErrHandlerNotFound
			break
		}

		result := q.processJob(&job, logger, handler)

		// If job was canceled while running a handler make sure it
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

		err = q.saveJobAndNotify(&job, logger, result)
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

func (q *queue) saveJobAndNotify(job *data.Job,
	logger log.Logger, result error) error {
	if err := q.db.Save(job); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	q.subsMtx.RLock()
	defer q.subsMtx.RUnlock()

	if subs, ok := q.subs[job.RelatedID]; ok {
		for _, v := range subs {
			go v.subFunc(job, result)
		}
	}

	return nil
}

func (q *queue) processJob(job *data.Job,
	logger log.Logger, handler Handler) error {

	tconf := q.typeConfig(job)

	logger.Info(fmt.Sprintf("processing job %s", job.Type))
	err := handler(job)

	if err == nil {
		job.Status = data.JobDone
		logger.Info(fmt.Sprintf("job %s is done", job.Type))
		return nil
	}

	if tconf.TryLimit != 0 {
		job.TryCount++
	}

	if job.TryCount >= tconf.TryLimit && tconf.TryLimit != 0 {
		job.Status = data.JobFailed
		logger.Error(fmt.Sprintf("job %s is failed", job.Type))
	} else {
		job.NotBefore = time.Now().Add(
			time.Duration(tconf.TryPeriod) * time.Millisecond)
		q.logger.Add("job", job).Warn(fmt.Sprintf(
			"retry for job %s(%s) scheduled to %s: %s",
			job.Type, job.ID, job.NotBefore.Format(time.RFC3339), err))
	}

	return err
}

func (q *queue) typeConfig(job *data.Job) TypeConfig {
	tconf := q.conf.TypeConfig
	if conf, ok := q.conf.Types[job.Type]; ok {
		tconf = conf
	}
	return tconf
}

// AddWithDataAndDelay is convenience method to add a job with given data
// and delay.
func AddWithDataAndDelay(q Queue, tx *reform.TX,
	jobType, relatedType, relatedID, creator string,
	jobData interface{}, delay time.Duration) error {
	data2, err := json.Marshal(jobData)
	if err != nil {
		return err
	}

	return q.Add(tx, &data.Job{
		Type:        jobType,
		RelatedType: relatedType,
		RelatedID:   relatedID,
		CreatedBy:   creator,
		Data:        data2,
		NotBefore:   time.Now().Add(delay),
	})
}

// AddWithData is convenience method to add a job with given data.
func AddWithData(q Queue, tx *reform.TX,
	jobType, relatedType, relatedID, creator string,
	jobData interface{}) error {
	return AddWithDataAndDelay(q, tx, jobType,
		relatedType, relatedID, creator, jobData, time.Duration(0))
}

// AddSimple is convenience method to add a job.
func AddSimple(q Queue, tx *reform.TX,
	jobType, relatedType, relatedID, creator string) error {
	return AddWithData(q, tx,
		jobType, relatedType, relatedID, creator, &struct{}{})
}

// AddWithDelay is convenience method to add a job with given data delay.
func AddWithDelay(q Queue, tx *reform.TX,
	jobType, relatedType, relatedID, creator string,
	delay time.Duration) error {
	return AddWithDataAndDelay(q, tx,
		jobType, relatedType, relatedID, creator, &struct{}{}, delay)
}

// SubFunc is a job result notification callback.
type SubFunc func(job *data.Job, result error)

func (q *queue) subscribe(relatedID, subID string, subFunc SubFunc) error {
	if subs, ok := q.subs[relatedID]; ok {
		for _, v := range subs {
			if v.subID == subID {
				return ErrSubscriptionExists
			}
		}
	}

	q.subs[relatedID] = append(q.subs[relatedID], subEntry{subID, subFunc})

	return nil
}

// Subscribe adds a subscription to job result notifications for given ids of
// related objects. SubID is used to distinguish between different
// subscriptions to a same related object.
func (q *queue) Subscribe(
	relatedIDs []string, subID string, subFunc SubFunc) error {
	q.subsMtx.Lock()
	defer q.subsMtx.Unlock()

	for i, v := range relatedIDs {
		if err := q.subscribe(v, subID, subFunc); err != nil {
			for j := 0; j < i; j++ {
				q.unsubscribe(relatedIDs[j], subID)
			}
			return err
		}
	}

	return nil
}

func (q *queue) unsubscribe(relatedID, subID string) error {
	if subs, ok := q.subs[relatedID]; ok {
		for i, v := range subs {
			if v.subID != subID {
				continue
			}

			subs := append(subs[:i], subs[i+1:]...)
			if subs != nil {
				q.subs[relatedID] = subs
			} else {
				delete(q.subs, relatedID)
			}

			return nil
		}
	}

	return ErrSubscriptionNotFound
}

// Unsubscribe removes a subscription from job result notifications for given
// ids of related objects. SubID is used to distinguish between different
// subscriptions to a same related object.
func (q *queue) Unsubscribe(relatedIDs []string, subID string) error {
	q.subsMtx.Lock()
	defer q.subsMtx.Unlock()

	var err error
	for _, v := range relatedIDs {
		if err2 := q.unsubscribe(v, subID); err == nil {
			err = err2
		}
	}

	return err
}
