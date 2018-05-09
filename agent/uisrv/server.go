package uisrv

import (
	"net/http"

	"github.com/ethereum/go-ethereum/ethclient"
	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth/contract"
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
	conf           *Config
	logger         *util.Logger
	db             *reform.DB
	ethClient      *ethclient.Client
	ptc            *contract.PrivatixTokenContract
	psc            *contract.PrivatixServiceContract
	pwdStorage     data.PWDGetSetter
	encryptKeyFunc data.EncryptedKeyFunc
	decryptKeyFunc data.ToPrivateKeyFunc
}

// NewServer creates a new agent server.
func NewServer(conf *Config,
	logger *util.Logger,
	db *reform.DB,
	ethClient *ethclient.Client,
	ptc *contract.PrivatixTokenContract,
	psc *contract.PrivatixServiceContract,
	pwdStorage data.PWDGetSetter) *Server {
	return &Server{conf, logger, db, ethClient, ptc, psc, pwdStorage,
		data.EncryptedKey, data.ToPrivateKey}
}

const (
	accountsPath        = "/ui/accounts/"
	authPath            = "/ui/auth"
	channelsPath        = "/ui/channels/"
	clientOfferingsPath = "/ui/client/offerings"
	endpointsPath       = "/ui/endpoints"
	offeringsPath       = "/ui/offerings/"
	productsPath        = "/ui/products"
	sessionsPath        = "/ui/sessions"
	settingsPath        = "/ui/settings"
	templatePath        = "/ui/templates"
)

// ListenAndServe starts a server.
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc(accountsPath, basicAuthMiddleware(s, s.handleAccounts))
	mux.HandleFunc(authPath, s.handleAuth)
	mux.HandleFunc(channelsPath, basicAuthMiddleware(s, s.handleChannels))
	mux.HandleFunc(clientOfferingsPath, basicAuthMiddleware(s, s.handleGetClientOfferings))
	mux.HandleFunc(endpointsPath, basicAuthMiddleware(s, s.handleGetEndpoints))
	mux.HandleFunc(offeringsPath, basicAuthMiddleware(s, s.handleOfferings))
	mux.HandleFunc(productsPath, basicAuthMiddleware(s, s.handleProducts))
	mux.HandleFunc(sessionsPath, basicAuthMiddleware(s, s.handleGetSessions))
	mux.HandleFunc(settingsPath, basicAuthMiddleware(s, s.handleSettings))
	mux.HandleFunc(templatePath, basicAuthMiddleware(s, s.handleTempaltes))
	mux.HandleFunc("/", s.pageNotFound)

	if s.conf.TLS != nil {
		return http.ListenAndServeTLS(
			s.conf.Addr, s.conf.TLS.CertFile, s.conf.TLS.KeyFile, mux)
	}

	return http.ListenAndServe(s.conf.Addr, mux)
}
