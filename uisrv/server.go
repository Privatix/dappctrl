package uisrv

import (
	"net/http"

	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util/log"
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
	logger         log.Logger
	db             *reform.DB
	dappRole       string
	queue          job.Queue
	pwdStorage     data.PWDGetSetter
	encryptKeyFunc data.EncryptedKeyFunc
	decryptKeyFunc data.ToPrivateKeyFunc
	pr             *proc.Processor
}

// NewServer creates a new agent server.
func NewServer(conf *Config,
	logger log.Logger,
	db *reform.DB,
	dappRole string,
	queue job.Queue,
	pwdStorage data.PWDGetSetter, pr *proc.Processor) *Server {
	return &Server{
		conf,
		logger,
		db,
		dappRole,
		queue,
		pwdStorage,
		data.EncryptedKey,
		data.ToPrivateKey,
		pr}
}

const (
	accountsPath        = "/accounts/"
	authPath            = "/auth"
	channelsPath        = "/channels/"
	clientChannelsPath  = "/client/channels/"
	clientOfferingsPath = "/client/offerings/"
	clientProductsPath  = "/client/products"
	endpointsPath       = "/endpoints"
	incomePath          = "/income"
	logsPath            = "/logs"
	offeringsPath       = "/offerings/"
	productsPath        = "/products"
	sessionsPath        = "/sessions"
	settingsPath        = "/settings"
	templatePath        = "/templates"
	transactionsPath    = "/transactions"
	usagePath           = "/usage"
	userRolePath        = "/userrole"
)

// ListenAndServe starts a server.
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()

	mux.HandleFunc(authPath, s.handleAuth)

	for _, item := range []struct {
		path    string
		handler http.HandlerFunc
	}{
		{
			path:    accountsPath,
			handler: s.handleAccounts,
		},
		{
			path:    channelsPath,
			handler: s.handleChannels,
		},
		{
			path:    clientChannelsPath,
			handler: s.handleClientChannels,
		},
		{
			path:    clientOfferingsPath,
			handler: s.handleClientOfferings,
		},
		{
			path:    clientProductsPath,
			handler: s.handleGetClientProducts,
		},
		{
			path:    endpointsPath,
			handler: s.handleGetEndpoints,
		},
		{
			path:    incomePath,
			handler: s.handleGetIncome,
		},
		{
			path:    logsPath,
			handler: s.handleGetLogs,
		},
		{
			path:    offeringsPath,
			handler: s.handleOfferings,
		},
		{
			path:    productsPath,
			handler: s.handleProducts,
		},
		{
			path:    sessionsPath,
			handler: s.handleGetSessions,
		},
		{
			path:    settingsPath,
			handler: s.handleSettings,
		},
		{
			path:    templatePath,
			handler: s.handleTempaltes,
		},
		{
			path:    transactionsPath,
			handler: s.handleTransactions,
		},
		{
			path:    usagePath,
			handler: s.handleGetUsage,
		},
		{
			path:    userRolePath,
			handler: s.handleGetUserRole,
		},
	} {
		mux.HandleFunc(item.path, basicAuthMiddleware(s, item.handler))
	}

	mux.HandleFunc("/", s.pageNotFound)

	if s.conf.TLS != nil {
		return http.ListenAndServeTLS(
			s.conf.Addr, s.conf.TLS.CertFile, s.conf.TLS.KeyFile, mux)
	}

	return http.ListenAndServe(s.conf.Addr, mux)
}
