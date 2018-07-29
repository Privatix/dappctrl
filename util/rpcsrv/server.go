package rpcsrv

import (
	"net/http"

	"github.com/ethereum/go-ethereum/rpc"
)

// TLSConfig is a TLS configuration.
type TLSConfig struct {
	CertFile string
	KeyFile  string
}

// Config is a RPC server configuration.
type Config struct {
	Addr           string
	AllowedOrigins []string
	TLS            *TLSConfig
}

// NewConfig creates a new configuration.
func NewConfig() *Config {
	return &Config{
		Addr:           "localhost:80",
		AllowedOrigins: []string{"*"},
		TLS:            nil,
	}
}

// Server is a RPC server which supports both HTTP and WS.
type Server struct {
	conf    *Config
	rpcsrv  *rpc.Server
	httpsrv *http.Server
}

// URL paths.
const (
	HTTPPath = "/http"
	WSPath   = "/ws"
)

// NewServer creates a new UI server.
func NewServer(conf *Config) (*Server, error) {
	rpcsrv := rpc.NewServer()

	mux := http.NewServeMux()
	mux.HandleFunc(HTTPPath, rpcsrv.ServeHTTP)
	mux.Handle(WSPath, rpcsrv.WebsocketHandler(conf.AllowedOrigins))

	httpsrv := &http.Server{
		Addr:    conf.Addr,
		Handler: mux,
	}

	return &Server{
		conf:    conf,
		rpcsrv:  rpcsrv,
		httpsrv: httpsrv,
	}, nil
}

// AddHandler registers a new RPC handler in a given namespace.
func (s *Server) AddHandler(namespace string, handler interface{}) error {
	return s.rpcsrv.RegisterName(namespace, handler)
}

// ListenAndServe starts to listen and to serve requests.
func (s *Server) ListenAndServe() error {
	if s.conf.TLS != nil {
		return s.httpsrv.ListenAndServeTLS(
			s.conf.TLS.CertFile, s.conf.TLS.KeyFile)
	}

	return s.httpsrv.ListenAndServe()
}

// Close immediately closes the server making ListenAndServe() to return.
func (s *Server) Close() error {
	return s.httpsrv.Close()
}
