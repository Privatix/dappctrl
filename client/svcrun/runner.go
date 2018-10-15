package svcrun

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/client/svcrun/exec"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util/log"
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

type startProcFunc func(
	name string, args []string, channel string) (*exec.Process, error)

type serviceRunner struct {
	conf      *Config
	logger    log.Logger
	db        *reform.DB
	pr        *proc.Processor
	startProc startProcFunc
	mtx       sync.Mutex
	procs     map[string]*exec.Process
	// The channel is only needed for tests.
	// It allows to receive a notification
	// about the end of a service launch.
	done chan bool
}

// NewServiceRunner creates a new service runner.
func NewServiceRunner(conf *Config, logger log.Logger,
	db *reform.DB, pr *proc.Processor) ServiceRunner {
	startProc := func(name string,
		args []string, channel string) (*exec.Process, error) {
		return exec.StartProcess(name, append(args, "-channel="+channel)...)
	}

	return &serviceRunner{
		conf:      conf,
		logger:    logger.Add("type", "svcrun.serviceRunner"),
		db:        db,
		pr:        pr,
		startProc: startProc,
		procs:     make(map[string]*exec.Process),
	}
}

// StopAll stops all the running services.
func (r *serviceRunner) StopAll() error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	var err error
	for _, v := range r.procs {
		err2 := v.Kill()
		if err == nil {
			err = err2
		}
	}

	return err
}

// Start starts a service associated with a given channel.
func (r *serviceRunner) Start(channel string) error {
	logger := r.logger.Add("method", "Start", "channel", channel)

	conf, key, err := r.getKey(channel)
	if err != nil {
		return err
	}

	r.mtx.Lock()
	defer r.mtx.Unlock()

	if _, ok := r.procs[key]; ok {
		return ErrAlreadyStarted
	}

	proc, err := r.startProc(conf.Name, conf.Args, channel)
	if err != nil {
		return err
	}

	logger.Warn("service adapter has started")

	r.procs[key] = proc

	go r.wait(channel, key, proc)

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

func (r *serviceRunner) wait(channel, key string, proc *exec.Process) {
	defer func() {
		select {
		case r.done <- true:
		default:
		}
	}()

	logger := r.logger.Add("method", "wait", "channel", channel)

	go func() {
		scanner := bufio.NewScanner(proc.Stderr)
		for scanner.Scan() {
			io.WriteString(
				os.Stderr, "dappvpn: "+scanner.Text()+"\n")
		}
		proc.Close()
	}()

	logger.Warn(fmt.Sprintf("service adapter has exited: %v", proc.Wait()))

	r.mtx.Lock()
	defer r.mtx.Unlock()

	delete(r.procs, key)

	_, err := r.pr.SuspendChannel(channel, data.JobServiceAdapter, false)
	if err != nil {
		logger.Add("error", err).Warn("failed to suspend channel")
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

	proc, ok := r.procs[key]
	if !ok {
		return ErrNotRunning
	}

	return proc.Kill()
}

// IsRunning returns whether a service for a given channel is running.
func (r *serviceRunner) IsRunning(channel string) (bool, error) {
	_, key, err := r.getKey(channel)
	if err != nil {
		return false, err
	}

	r.mtx.Lock()
	defer r.mtx.Unlock()

	_, ok := r.procs[key]

	return ok, nil
}
