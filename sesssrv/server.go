package sesssrv

import (
	"net/http"

	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/util"
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
	conf *Config
	db   *reform.DB
}

// Service API paths.
const (
	PathAuth   = "/session/auth"
	PathStart  = "/session/start"
	PathStop   = "/session/stop"
	PathUpdate = "/session/update"

	PathProductConfig = "/product/config"
)

// NewServer creates a new session server.
func NewServer(conf *Config, logger *util.Logger, db *reform.DB) *Server {
	s := &Server{
		Server: srv.NewServer(conf.Config, logger),
		conf:   conf,
		db:     db,
	}

	modifyHandler := func(h srv.HandlerFunc) srv.HandlerFunc {
		h = s.RequireBasicAuth(h, s.authProduct)
		h = s.RequireHTTPMethods(h, http.MethodPost)
		return h
	}

	s.HandleFunc(PathAuth, modifyHandler(s.handleAuth))
	s.HandleFunc(PathStart, modifyHandler(s.handleStart))
	s.HandleFunc(PathStop, modifyHandler(s.handleStop))
	s.HandleFunc(PathUpdate, modifyHandler(s.handleUpdate))
	s.HandleFunc(PathProductConfig, modifyHandler(s.handleProductConfig))

	return s
}
