package uisrv

import (
	"net/http"

	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/eth/contract"

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
	conf      *Config
	logger    *util.Logger
	db        *reform.DB
	ethClient *eth.EthereumClient
	ptc       *contract.PrivatixTokenContract
	psc       *contract.PrivatixServiceContract
}

// NewServer creates a new agent server.
func NewServer(conf *Config,
	logger *util.Logger,
	db *reform.DB,
	ethClient *eth.EthereumClient,
	ptc *contract.PrivatixTokenContract,
	psc *contract.PrivatixServiceContract) *Server {
	return &Server{conf, logger, db, ethClient, ptc, psc}
}

const (
	accountsPath  = "/ui/accounts/"
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
	mux.HandleFunc(accountsPath, basicAuthMiddlewareFunc(s, s.handleAccounts))
	mux.HandleFunc(authPath, s.handleAuth)
	mux.HandleFunc(channelsPath, basicAuthMiddlewareFunc(s, s.handleChannels))
	mux.HandleFunc(endpointsPath, basicAuthMiddlewareFunc(s, s.handleGetEndpoints))
	mux.HandleFunc(offeringsPath, basicAuthMiddlewareFunc(s, s.handleOfferings))
	mux.HandleFunc(productsPath, basicAuthMiddlewareFunc(s, s.handleProducts))
	mux.HandleFunc(sessionsPath, basicAuthMiddlewareFunc(s, s.handleGetSessions))
	mux.HandleFunc(settingsPath, basicAuthMiddlewareFunc(s, s.handleSettings))
	mux.HandleFunc(templatePath, basicAuthMiddlewareFunc(s, s.handleTempaltes))
	mux.HandleFunc("/", s.pageNotFound)

	if s.conf.TLS != nil {
		return http.ListenAndServeTLS(
			s.conf.Addr, s.conf.TLS.CertFile, s.conf.TLS.KeyFile, mux)
	}

	return http.ListenAndServe(s.conf.Addr, mux)
}
