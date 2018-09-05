// Package connector implements standard methods
// for communicating with dappctrl.
package connector

import (
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/util/srv"
)

// Config is a connector configuration.
type Config struct {
	*srv.Config
	Username string
	Password string
}

// Connector defines the methods for interacting with a dappctrl.
type Connector interface {
	// AuthSession authenticates client session via dappctrl.
	AuthSession(args *sesssrv.AuthArgs) error
	// StartSession reports session start.
	StartSession(args *sesssrv.StartArgs) error
	// StopSession reports session stop.
	StopSession(args *sesssrv.StopArgs) error
	// UpdateSessionUsage reports last usage update.
	UpdateSessionUsage(args *sesssrv.UpdateArgs) error
	// SetupProductConfiguration send common server configuration
	// used by client access message.
	SetupProductConfiguration(args *sesssrv.ProductArgs) error
	// GetEndpointMessage returns endpoint message by channel identificator.
	GetEndpointMessage(
		args *sesssrv.EndpointMsgArgs) (*data.Endpoint, error)
}

type cntr struct {
	config *Config
	logger log.Logger
}

// DefaultConfig is a default connector config.
func DefaultConfig() *Config {
	return &Config{Config: srv.NewConfig()}
}

// NewConnector implements standard connector for communicating with dappctrl.
func NewConnector(config *Config, logger log.Logger) Connector {
	return &cntr{
		config: config,
		logger: logger,
	}
}

// AuthSession sends a request for session authentication.
func (c *cntr) AuthSession(args *sesssrv.AuthArgs) error {
	return sesssrv.Post(c.config.Config, c.logger, c.config.Username,
		c.config.Password, sesssrv.PathAuth, args, nil)
}

// StartSession sends a request for session start.
func (c *cntr) StartSession(args *sesssrv.StartArgs) error {
	return sesssrv.Post(c.config.Config, c.logger, c.config.Username,
		c.config.Password, sesssrv.PathStart, args, nil)
}

// StopSession sends a request for session stop.
func (c *cntr) StopSession(args *sesssrv.StopArgs) error {
	return sesssrv.Post(c.config.Config, c.logger, c.config.Username,
		c.config.Password, sesssrv.PathStop, args, nil)
}

// UpdateSessionUsage sends a request to update
// a information on the use of session.
func (c *cntr) UpdateSessionUsage(args *sesssrv.UpdateArgs) error {
	return sesssrv.Post(c.config.Config, c.logger, c.config.Username,
		c.config.Password, sesssrv.PathUpdate, args, nil)
}

// SetupProductConfiguration  sends a request to update product configuration.
func (c *cntr) SetupProductConfiguration(args *sesssrv.ProductArgs) error {
	return sesssrv.Post(c.config.Config, c.logger, c.config.Username,
		c.config.Password, sesssrv.PathProductConfig, args, nil)
}

// GetEndpointMessage returns endpoint message by channel identificator.
func (c *cntr) GetEndpointMessage(
	args *sesssrv.EndpointMsgArgs) (*data.Endpoint, error) {
	var endpoint *data.Endpoint
	err := sesssrv.Post(c.config.Config, c.logger, c.config.Username,
		c.config.Password, sesssrv.PathProductConfig, args, &endpoint)
	return endpoint, err
}
