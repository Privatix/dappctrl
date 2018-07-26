package config

import (
	"time"

	"github.com/privatix/dappctrl/svc/dappvpn/mon"
	"github.com/privatix/dappctrl/svc/dappvpn/msg"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/srv"
)

type ovpnConfig struct {
	Name       string        // Name of OvenVPN executable.
	Args       []string      // Extra arguments for OpenVPN executable.
	ConfigRoot string        // Root path for OpenVPN channel configs.
	StartDelay time.Duration // Delay to ensure OpenVPN is ready, in milliseconds.
}

type serverConfig struct {
	*srv.Config
	Password string
	Username string
}

// Config is dappvpn configuration.
type Config struct {
	ChannelDir string // Directory for common-name -> channel mappings.
	Log        *util.LogConfig
	Monitor    *mon.Config
	OpenVPN    *ovpnConfig // OpenVPN settings for client mode.
	Pusher     *msg.Config
	Server     *serverConfig
}

// NewConfig creates default dappvpn configuration.
func NewConfig() *Config {
	return &Config{
		ChannelDir: ".",
		Log:        util.NewLogConfig(),
		Monitor:    mon.NewConfig(),
		OpenVPN: &ovpnConfig{
			Name:       "openvpn",
			ConfigRoot: "/etc/openvpn/config",
			StartDelay: 1000,
		},
		Pusher: msg.NewConfig(),
		Server: &serverConfig{Config: srv.NewConfig()},
	}
}
