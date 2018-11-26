package looper

import (
	"context"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util/log"
)

// Config is a looper configuration.
type Config struct {
	AutoOfferingPopUpTimeout uint64 // in seconds
}

// NewConfig creates default looper configuration.
func NewConfig() *Config {
	return &Config{}
}

var (
	// BlockTime is estimated time needed to create a block in Ethereum.
	BlockTime = time.Second * 15
)

// Loop performs a function that creates new jobs with a certain frequency.
// If at the moment there is already a similar active job, then such job is
// ignored. The function works correctly only with jobs for which duplicates
// are allowed.
func Loop(ctx context.Context, logger log.Logger, db *reform.DB,
	queue job.Queue, duration time.Duration, f func() []*data.Job) {
	tik := time.NewTicker(duration)
	logger = logger.Add("method", "Loop")

	go loop(ctx, tik, db, queue, f, logger, nil)
}

func loop(ctx context.Context, tik *time.Ticker, db *reform.DB, queue job.Queue,
	f func() []*data.Job, logger log.Logger, pulse chan struct{}) {
	for {
		select {
		case <-tik.C:
			logger.Debug("new iteration")
			for _, j := range f() {
				res, err := db.SelectAllFrom(data.JobTable,
					`WHERE related_id = $1
						AND type = $2
						AND status = $3`,
					j.RelatedID, j.Type, data.JobActive)
				if err != nil {
					logger.Error(err.Error())
					break
				}

				if len(res) > 0 {
					continue
				}

				err = queue.Add(nil, j)
				if err != nil {
					logger.Error(err.Error())
					break
				}
			}
			if pulse != nil {
				pulse <- struct{}{}
			}
		case <-ctx.Done():
			return
		}
	}
}
