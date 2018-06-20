package svcrun

import (
	"errors"
	"os/exec"
	"sync"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util"
)

// ServiceRunner errors.
var (
	ErrAlreadyStarted = errors.New("service already running")
	ErrUnknownService = errors.New("unknown service type")
	ErrNotRunning     = errors.New("not running")
)

// ServiceRunner starts and stops services.
type ServiceRunner interface {
	Start(channel string) error
	IsRunning(channel string) (bool, error)
	Stop(channel string) error
	StopAll() error
}

// ServiceConfig is a service-specific configuration.
type ServiceConfig struct {
	Name   string   // Name of client adapter executable.
	Args   []string // Client adapter arguments.
	Single bool     // Whether to forbid multiple service instances.
}

// Config is a service runner configuration.
type Config struct {
	Services map[string]ServiceConfig
}

// NewConfig creates a new service runner configuration.
func NewConfig() *Config {
	return &Config{
		Services: map[string]ServiceConfig{},
	}
}

type newCmdFunc func(name string, args []string, channel string) *exec.Cmd

type serviceRunner struct {
	conf   *Config
	logger *util.Logger
	db     *reform.DB
	queue  *job.Queue
	newCmd newCmdFunc
	mtx    sync.Mutex
	cmds   map[string]*exec.Cmd
}

// NewServiceRunner creates a new service runner.
func NewServiceRunner(conf *Config, logger *util.Logger,
	db *reform.DB, queue *job.Queue) ServiceRunner {
	newCmd := func(name string, args []string, channel string) *exec.Cmd {
		return exec.Command(name, append(args, "-channel="+channel)...)
	}

	return &serviceRunner{
		conf:   conf,
		logger: logger,
		db:     db,
		queue:  queue,
		newCmd: newCmd,
		cmds:   make(map[string]*exec.Cmd),
	}
}

// StopAll stops all the running services.
func (r *serviceRunner) StopAll() error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	var err error
	for _, v := range r.cmds {
		if err2 := v.Process.Kill(); err == nil {
			err = err2
		}
	}

	return err
}

// Start starts a service associated with a given channel.
func (r *serviceRunner) Start(channel string) error {
	conf, key, err := r.getKey(channel)
	if err != nil {
		return err
	}

	r.mtx.Lock()
	defer r.mtx.Unlock()

	if _, ok := r.cmds[key]; ok {
		return ErrAlreadyStarted
	}

	cmd := r.newCmd(conf.Name, conf.Args, channel)
	if err := cmd.Start(); err != nil {
		return err
	}

	r.logger.Warn("service adapter for channel %s has started", channel)

	r.cmds[key] = cmd

	go r.wait(channel, key, cmd)

	return nil
}

func (r *serviceRunner) getKey(channel string) (*ServiceConfig, string, error) {
	var ch data.Channel
	err := data.FindByPrimaryKeyTo(r.db.Querier, &ch, channel)
	if err != nil {
		return nil, "", err
	}

	var offer data.Offering
	err = data.FindByPrimaryKeyTo(r.db.Querier, &offer, ch.Offering)
	if err != nil {
		return nil, "", err
	}

	conf, ok := r.conf.Services[offer.ServiceName]
	if !ok {
		return nil, "", ErrUnknownService
	}

	if conf.Single {
		return &conf, offer.ServiceName, nil
	}

	return &conf, channel, nil
}

func (r *serviceRunner) wait(channel, key string, cmd *exec.Cmd) {
	r.logger.Warn("service adapter for channel %s has exited: %v",
		channel, cmd.Wait())

	r.mtx.Lock()
	defer r.mtx.Unlock()

	delete(r.cmds, key)

	if err := r.queue.Add(&data.Job{
		Type:        data.JobClientPreServiceSuspend,
		RelatedType: data.JobChannel,
		RelatedID:   channel,
		CreatedBy:   data.JobServiceAdapter,
		Data:        []byte("{}"),
	}); err != nil {
		r.logger.Error("failed to add a job to the queue: %s", err)
	}
}

// Stop stops an already started service.
func (r *serviceRunner) Stop(channel string) error {
	_, key, err := r.getKey(channel)
	if err != nil {
		return err
	}

	r.mtx.Lock()
	defer r.mtx.Unlock()

	cmd, ok := r.cmds[key]
	if !ok {
		return ErrNotRunning
	}

	return cmd.Process.Kill()
}

// IsRunning returns whether a service for a given channel is running.
func (r *serviceRunner) IsRunning(channel string) (bool, error) {
	_, key, err := r.getKey(channel)
	if err != nil {
		return false, err
	}

	r.mtx.Lock()
	defer r.mtx.Unlock()

	_, ok := r.cmds[key]

	return ok, nil
}
