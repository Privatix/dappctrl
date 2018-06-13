package uisrv

import (
	"net/http"

	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
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
	Addr           string
	TLS            *TLSConfig
	EthCallTimeout uint // In seconds.
}

// NewConfig creates a default server configuration.
func NewConfig() *Config {
	return &Config{
		EthCallTimeout: 5,
	}
}

// Server is agent api server.
type Server struct {
	conf           *Config
	logger         *util.Logger
	db             *reform.DB
	queue          *job.Queue
	pwdStorage     data.PWDGetSetter
	encryptKeyFunc data.EncryptedKeyFunc
	decryptKeyFunc data.ToPrivateKeyFunc
}

// NewServer creates a new agent server.
func NewServer(conf *Config,
	logger *util.Logger,
	db *reform.DB,
	queue *job.Queue,
	pwdStorage data.PWDGetSetter) *Server {
	return &Server{
		conf,
		logger,
		db,
		queue,
		pwdStorage,
		data.EncryptedKey,
		data.ToPrivateKey}
}

const (
	accountsPath        = "/accounts/"
	authPath            = "/auth"
	channelsPath        = "/channels/"
	clientChannelsPath  = "/client/channels/"
	clientOfferingsPath = "/client/offerings"
	clientProductsPath  = "/client/products"
	endpointsPath       = "/endpoints"
	incomePath          = "/income"
	offeringsPath       = "/offerings/"
	productsPath        = "/products"
	sessionsPath        = "/sessions"
	settingsPath        = "/settings"
	templatePath        = "/templates"
	transactionsPath    = "/transactions"
	usagePath           = "/usage"
)

// ListenAndServe starts a server.
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc(accountsPath, basicAuthMiddleware(s, s.handleAccounts))
	mux.HandleFunc(authPath, s.handleAuth)
	mux.HandleFunc(channelsPath, basicAuthMiddleware(s, s.handleChannels))
	mux.HandleFunc(clientChannelsPath,
		basicAuthMiddleware(s, s.handleGetClientChannels))
	mux.HandleFunc(clientOfferingsPath,
		basicAuthMiddleware(s, s.handleGetClientOfferings))
	mux.HandleFunc(clientProductsPath,
		basicAuthMiddleware(s, s.handleGetClientProducts))
	mux.HandleFunc(endpointsPath, basicAuthMiddleware(s, s.handleGetEndpoints))
	mux.HandleFunc(incomePath, basicAuthMiddleware(s, s.handleGetIncome))
	mux.HandleFunc(offeringsPath, basicAuthMiddleware(s, s.handleOfferings))
	mux.HandleFunc(productsPath, basicAuthMiddleware(s, s.handleProducts))
	mux.HandleFunc(sessionsPath, basicAuthMiddleware(s, s.handleGetSessions))
	mux.HandleFunc(settingsPath, basicAuthMiddleware(s, s.handleSettings))
	mux.HandleFunc(templatePath, basicAuthMiddleware(s, s.handleTempaltes))
	mux.HandleFunc(transactionsPath, basicAuthMiddleware(s, s.handleTransactions))
	mux.HandleFunc(usagePath, basicAuthMiddleware(s, s.handleGetUsage))
	mux.HandleFunc("/", s.pageNotFound)

	if s.conf.TLS != nil {
		return http.ListenAndServeTLS(
			s.conf.Addr, s.conf.TLS.CertFile, s.conf.TLS.KeyFile, mux)
	}

	return http.ListenAndServe(s.conf.Addr, mux)
}
