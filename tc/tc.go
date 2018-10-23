package tc

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/privatix/dappctrl/util/log"
)

// TrafficControl is a traffic control utility.
type TrafficControl struct {
	conf   *Config
	logger log.Logger
}

// NewTrafficControl creates a new TrafficControl instance.
func NewTrafficControl(conf *Config, logger log.Logger) *TrafficControl {
	return &TrafficControl{conf, logger.Add("type", "tc/TrafficControl")}
}

func (tc *TrafficControl) run(logger log.Logger,
	name string, args ...string) (stdout string, err error) {
	logger = logger.Add("cmd", name, "args", args)

	logger.Info("run command")

	cmd := exec.Command(name, args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()

	lines := strings.TrimSpace(stderr.String())
	for _, line := range strings.Split(lines, "\n") {
		if len(line) != 0 {
			logger.Warn(line)
		}
	}

	return string(out), err
}
