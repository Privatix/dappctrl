package main

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/client/svcrun"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/svc/dappvpn/mon"
	"github.com/privatix/dappctrl/util"
)

const (
	service = "VPN client"
)

// managementPortForClient returns OpenVpn management
// interface port from configuration.
func managementPortForClient(conf *svcrun.Config, logger *util.Logger,
	db *reform.Querier) uint16 {
	// it must only work for a client.
	if agent, err := data.IsAgent(db); err != nil || agent {
		return 0
	}

	// finds the VPN client service in the service runner configuration.
	serviceConfig, ok := conf.Services[service]
	if !ok {
		logger.Info("VPN client service is missing in" +
			" the configuration")
		return 0
	}

	prefix := "-config="

	var dappvpnConfigFile string

	// finds dappvpn configuration file path.
	for _, str := range serviceConfig.Args {
		if strings.HasPrefix(str, prefix) {
			dappvpnConfigFile = strings.TrimPrefix(str, prefix)
			break
		}
	}

	if dappvpnConfigFile == "" {
		logger.Warn("configuration for dappvpn not found")
		return 0
	}

	// reads dappvpn configuration file.
	configData, err := ioutil.ReadFile(dappvpnConfigFile)
	if err != nil {
		logger.Warn("can not read configuration"+
			" for dappvpn: %s", err)
		return 0
	}

	dappvpnConfig := struct{ Monitor *mon.Config }{
		Monitor: mon.NewConfig(),
	}

	if err := json.Unmarshal(configData, &dappvpnConfig); err != nil {
		logger.Warn("can not unmarshal dappvpn configuration: %s", err)
		return 0
	}

	// reads OpenVpn management interface address from configuration.
	params := strings.Split(dappvpnConfig.Monitor.Addr, ":")
	if len(params) != 2 {
		logger.Warn("address for OpenVpn management" +
			" interface is wrong")
		return 0
	}

	// reads OpenVpn management interface server port from configuration.
	port, err := strconv.ParseUint(params[1], 10, 16)
	if err != nil {
		logger.Warn("port for OpenVpn management"+
			" interface is wrong: %s", err)
		return 0
	}
	return uint16(port)
}
