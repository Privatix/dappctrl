package uisrv

import (
	"net/http"

	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/util"
)

// ActionPayload is a body format for action requests.
type ActionPayload struct {
	Action string `json:"action"`
}

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
	authPath      = "/ui/auth"
	channelsPath  = "/ui/channels/"
	endpointsPath = "/ui/endpoints"
	offeringsPath = "/ui/offerings/"
	productsPath  = "/ui/products"
	sessionsPath  = "/ui/sessions"
	settingsPath  = "/ui/settings"
	templatePath  = "/ui/templates"
)

// ListenAndServe starts a server.
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc(authPath, s.handleAuth)
	mux.HandleFunc(channelsPath, s.handleChannels)
	mux.HandleFunc(endpointsPath, basicAuthMiddlewareFunc(s, s.handleGetEndpoints))
	mux.HandleFunc(offeringsPath, s.handleOfferings)
	mux.HandleFunc(productsPath, s.handleProducts)
	mux.HandleFunc(sessionsPath, basicAuthMiddlewareFunc(s, s.handleGetSessions))
	mux.HandleFunc(settingsPath, s.handleSettings)
	mux.HandleFunc(templatePath, s.handleTempaltes)
	mux.HandleFunc("/", s.pageNotFound)

	if s.conf.TLS != nil {
		return http.ListenAndServeTLS(
			s.conf.Addr, s.conf.TLS.CertFile, s.conf.TLS.KeyFile, mux)
	}

	return http.ListenAndServe(s.conf.Addr, mux)
}
