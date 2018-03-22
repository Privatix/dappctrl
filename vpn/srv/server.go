package srv

import (
	"net/http"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/util"
)

// TLSConfig is a tls configuration.
type TLSConfig struct {
	CertFile string
	KeyFile  string
}

// Config is a configuration for VPN session server.
type Config struct {
	Addr string
	TLS  *TLSConfig
}

// NewConfig creates a default configuration for VPN session server.
func NewConfig() *Config {
	return &Config{
		Addr: "localhost:8000",
		TLS:  nil,
	}
}

// Server is a VPN session server.
type Server struct {
	conf   *Config
	logger *util.Logger
	db     *reform.DB
	srv    http.Server
}

// NewServer creates a new VPN session server.
func NewServer(conf *Config, logger *util.Logger, db *reform.DB) *Server {
	return &Server{conf, logger, db, http.Server{Addr: conf.Addr}}
}

// VPN session API paths.
const (
	PathAuth  = "/auth"
	PathStart = "/start"
	PathStop  = "/stop"
)

// ListenAndServe starts to listen and to serve requests.
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc(PathAuth, s.handleAuth)
	mux.HandleFunc(PathStart, s.handleStart)
	mux.HandleFunc(PathStop, s.handleStop)
	s.srv.Handler = mux

	if s.conf.TLS != nil {
		return s.srv.ListenAndServeTLS(s.conf.TLS.CertFile, s.conf.TLS.KeyFile)
	}

	return s.srv.ListenAndServe()
}

// Close immediately closes the server making ListenAndServe() to return.
func (s *Server) Close() error {
	return s.srv.Close()
}
