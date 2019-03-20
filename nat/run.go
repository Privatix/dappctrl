package nat

import (
	"context"
	"fmt"
	"time"

	"github.com/rdegges/go-ipify"

	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

// Run matches and opens external network ports.
func Run(ctx context.Context, conf *Config,
	logger log.Logger, ports []uint16) {
	if conf.Mechanism == "" {
		logger.Debug("traversal NAT is not needed.")
		return
	}

	service, err := Parse(conf)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	for k := range ports {
		// Remove old rules.
		service.DeleteMapping("tcp", int(ports[k]), int(ports[k]))

		name := fmt.Sprintf("service-%d", k)
		if err := Map(ctx, conf, logger, service, "tcp",
			int(ports[k]), int(ports[k]), name); err != nil {
			msg := fmt.Sprintf("failed to add port"+
				" mapping to %d port", ports[k])
			logger.Warn(msg)
			return
		}
	}

	extIP, err := ipify.GetIp()
	if err != nil {
		logger.Warn("failed to determine" +
			" a external ip address, error: " + err.Error())
		return
	}

	logger = logger.Add("externalIP", extIP)

	timeout := time.Duration(conf.CheckTimeout) * time.Millisecond

	checkSrv := func(port uint16) {
		if util.CheckConnection(
			"tcp", extIP, int(port), timeout) {
			logger.Info(fmt.Sprintf("port %d is available"+
				" on the Internet", port))
			return
		}
		logger.Warn(fmt.Sprintf("port %d is not available"+
			" on the Internet", port))
	}

	for k := range ports {
		checkSrv(ports[k])
	}
}
