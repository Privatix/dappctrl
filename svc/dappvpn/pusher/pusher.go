package pusher

import (
	"context"

	c "github.com/privatix/dappctrl/messages/ept/config"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/srv"
)

// Config for pushing OpenVpn configuration.
// ExportConfigKeys - list of parameters that are exported to
// the OpenVpn client configuration from the OpenVpn server configuration.
// ConfigPath - absolute path to OpenVpn server configuration file.
// CaCertPath - absolute path to Ca certificate file
// Pushed - if the configuration is passed to the Session server
// then this parameter is true.
// TimeOut - pause between attempts
// to send a configuration to the Session server.
type Config struct {
	ExportConfigKeys []string
	ConfigPath       string
	CaCertPath       string
	Pushed           bool
	TimeOut          int64
}

// Collect collects the required parameters.
type Collect struct {
	Config   *Config
	Server   *srv.Config
	Username string
	Password string
	Logger   *util.Logger
}

// NewConfig for create empty config.
func NewConfig() *Config {
	return &Config{}
}

// NewCollect for create new Collect object.
func NewCollect(conf *Config, srv *srv.Config, user, pass string,
	logger *util.Logger) *Collect {
	return &Collect{
		Config:   conf,
		Server:   srv,
		Username: user,
		Password: pass,
		Logger:   logger,
	}
}

func push(ctx context.Context, username, pass string, config *Config,
	srvConfig *srv.Config, logger *util.Logger) error {
	req := c.NewPushConfigReq(username, pass, config.ConfigPath,
		config.CaCertPath, config.ExportConfigKeys, config.TimeOut)

	return c.PushConfig(ctx, srvConfig, logger, req)
}

// PushConfig send the OpenVpn configuration to Session server.
func PushConfig(ctx context.Context, c *Collect) error {
	return push(ctx, c.Username, c.Password, c.Config, c.Server, c.Logger)
}
