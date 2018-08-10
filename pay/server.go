package pay

import (
	"net/http"

	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/util/srv"
)

// Config is a configuration for a pay server.
type Config struct {
	*srv.Config
}

// NewConfig creates a default pay server configuration.
func NewConfig() *Config {
	return &Config{
		Config: srv.NewConfig(),
	}
}

// Server is a pay server.
type Server struct {
	*srv.Server

	db     *reform.DB
	logger log.Logger
}

const payPath = "/v1/pmtChannel/pay"

// NewServer creates a new pay server.
func NewServer(conf *Config, logger log.Logger, db *reform.DB) *Server {
	s := &Server{
		Server: srv.NewServer(conf.Config),
		logger: logger.Add("type", "pay.Server"),
		db:     db,
	}

	s.HandleFunc(payPath,
		s.RequireHTTPMethods(s.logger, s.handlePay, http.MethodPost))

	return s
}
