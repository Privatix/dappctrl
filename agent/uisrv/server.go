package uisrv

import (
	"net/http"

	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/util"
)

// TLSConfig is a tls configuration.
type TLSConfig struct {
	CertFile string
	KeyFile  string
}

// Config is a configuration for a agent server.
type Config struct {
	Addr string
	TLS  *TLSConfig
}

// Server is agent api server.
type Server struct {
	conf   *Config
	logger *util.Logger
	db     *reform.DB
}

// NewServer creates a new agent server.
func NewServer(conf *Config, logger *util.Logger, db *reform.DB) *Server {
	return &Server{conf, logger, db}
}

const (
	channelsPath  = "/ui/channels/"
	endpointsPath = "/ui/endpoints"
	offeringsPath = "/ui/offerings/"
	productsPath  = "/ui/products"
	sessionsPath  = "/ui/sessions"
	settingsPath  = "/ui/settings"
	templatePath  = "/ui/templates"
	notFoundPath  = "/"
)

// ListenAndServe starts a server.
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc(channelsPath, s.handleChannels)
	mux.HandleFunc(endpointsPath, s.handleGetEndpoints)
	mux.HandleFunc(offeringsPath, s.handleOfferings)
	mux.HandleFunc(productsPath, s.handleProducts)
	mux.HandleFunc(sessionsPath, s.handleSessions)
	mux.HandleFunc(settingsPath, s.handleSettings)
	mux.HandleFunc(templatePath, s.handleTempaltes)
	mux.HandleFunc(notFoundPath, s.notFoundHandler)

	if s.conf.TLS != nil {
		return http.ListenAndServeTLS(
			s.conf.Addr, s.conf.TLS.CertFile, s.conf.TLS.KeyFile, mux)
	}

	return http.ListenAndServe(s.conf.Addr, mux)
}
