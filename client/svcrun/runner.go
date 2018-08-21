package svcrun

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util"
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
	pr     *proc.Processor
	newCmd newCmdFunc
	mtx    sync.Mutex
	cmds   map[string]*exec.Cmd
}

// NewServiceRunner creates a new service runner.
func NewServiceRunner(conf *Config, logger *util.Logger,
	db *reform.DB, pr *proc.Processor) ServiceRunner {
	newCmd := func(name string, args []string, channel string) *exec.Cmd {
		return exec.Command(name, append(args, "-channel="+channel)...)
	}

	return &serviceRunner{
		conf:   conf,
		logger: logger,
		db:     db,
		pr:     pr,
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
		err2 := syscall.Kill(-v.Process.Pid, syscall.SIGKILL)
		if err == nil {
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

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	r.logger.Warn("service adapter for channel %s has started", channel)

	r.cmds[key] = cmd

	go r.wait(channel, key, cmd, stderr)

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

func (r *serviceRunner) wait(
	channel, key string, cmd *exec.Cmd, stderr io.ReadCloser) {
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			io.WriteString(
				os.Stderr, "dappvpn: "+scanner.Text()+"\n")
		}
		stderr.Close()
	}()

	r.logger.Warn("service adapter for channel %s has exited: %v",
		channel, cmd.Wait())

	r.mtx.Lock()
	defer r.mtx.Unlock()

	delete(r.cmds, key)

	_, err := r.pr.SuspendChannel(channel, data.JobServiceAdapter, false)
	if err != nil {
		r.logger.Warn("failed to suspend channel: %s", err)
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

	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
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
