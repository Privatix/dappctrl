package ui

import (
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/util/rpcsrv"
)

// Config is a handler configuration.
type Config struct {
	*rpcsrv.Config
}

// NewConfig creates a new handler configuration.
func NewConfig() *Config {
	return &Config{
		Config: rpcsrv.NewConfig(),
	}
}

// Handler is an UI RPC handler.
type Handler struct {
	conf   *Config
	logger log.Logger
	db     *reform.DB
	queue  job.Queue
}

// NewHandler creates a new handler.
func NewHandler(conf *Config,
	logger log.Logger, db *reform.DB, queue job.Queue) *Handler {
	logger = logger.Add("type", "uisrv.Handler")
	return &Handler{
		conf:   conf,
		logger: logger,
		db:     db,
		queue:  queue,
	}
}
