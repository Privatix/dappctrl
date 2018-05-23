package pusher

import (
	"context"

	c "github.com/privatix/dappctrl/ept/config"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/srv"
)

// Config for Pusher object
type Config struct {
	ExportConfigParams []string
	ConfigPath         string
	CaCertPath         string
	Pushed             bool
	TimeOut            int64
}

// Pusher sends the OpenVpn configuration to sessrv
type Pusher struct {
	c       *Config
	srvConf *srv.Config
	logger  *util.Logger
}

// NewPusher creates a new Pusher object
func NewPusher(config *Config, srvConfig *srv.Config,
	logger *util.Logger) *Pusher {
	return &Pusher{
		c:       config,
		srvConf: srvConfig,
		logger:  logger,
	}
}

// Push send the OpenVpn configuration to sessrv
func (p *Pusher) Push(ctx context.Context, username, pass string) error {
	req := c.NewPushConfigReq(username, pass, p.c.ConfigPath,
		p.c.CaCertPath, p.c.ExportConfigParams, p.c.TimeOut)

	return c.PushConfig(ctx, p.srvConf, p.logger, req)
}

// Context creates a new context with cancel function
func Context() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}
