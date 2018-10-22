package sesssrv

import (
	"net/http"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/country"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/util/srv"
)

// Config is a session server configuration.
type Config struct {
	*srv.Config
}

// NewConfig creates a default session server configuration.
func NewConfig() *Config {
	return &Config{
		Config: srv.NewConfig(),
	}
}

// Server is a service session server.
type Server struct {
	*srv.Server
	conf        *Config
	countryConf *country.Config
	db          *reform.DB
	logger      log.Logger
	queue       job.Queue
}

// Service API paths.
const (
	PathAuth   = "/session/auth"
	PathStart  = "/session/start"
	PathStop   = "/session/stop"
	PathUpdate = "/session/update"

	PathProductConfig = "/product/config"
	PathEndpointMsg   = "/endpoint/message"
)

// NewServer creates a new session server.
func NewServer(conf *Config, logger log.Logger, db *reform.DB,
	countryConf *country.Config, queue job.Queue) *Server {
	s := &Server{
		Server:      srv.NewServer(conf.Config),
		conf:        conf,
		db:          db,
		logger:      logger.Add("type", "sesssrv.Server"),
		countryConf: countryConf,
		queue:       queue,
	}

	modifyHandler := func(h srv.HandlerFunc) srv.HandlerFunc {
		h = s.RequireBasicAuth(s.logger, h, s.authProduct)
		h = s.RequireHTTPMethods(s.logger, h, http.MethodPost)
		return h
	}

	s.HandleFunc(PathAuth, modifyHandler(s.handleAuth))
	s.HandleFunc(PathStart, modifyHandler(s.handleStart))
	s.HandleFunc(PathStop, modifyHandler(s.handleStop))
	s.HandleFunc(PathUpdate, modifyHandler(s.handleUpdate))
	s.HandleFunc(PathEndpointMsg, modifyHandler(s.handleEndpointMsg))
	s.HandleFunc(PathProductConfig, modifyHandler(s.handleProductConfig))

	return s
}
