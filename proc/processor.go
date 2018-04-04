package proc

import (
	"errors"

	"github.com/privatix/dappctrl/job/queue"
)

// Config is processor configuration.
type Config struct {
}

// NewConfig creates a default processor configuration.
func NewConfig() *Config {
	return &Config{}
}

// Processor encapsulates a set of top-level business logic routines.
type Processor struct {
	conf  *Config
	queue *queue.Queue
}

// NewProcessor creates a new processor.
func NewProcessor(
	conf *Config, queue *queue.Queue) *Processor {
	return &Processor{
		conf:  conf,
		queue: queue,
	}
}

// Processor-specific errors.
var (
	ErrBadServiceStatus = errors.New("bad service status")
	ErrActiveJobsExist  = errors.New("active jobs exist")
	ErrSameJobExists    = errors.New("same jobs exists")
)
