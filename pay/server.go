package pay

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

// Config is a configuration for a payment server.
type Config struct {
	Addr string
	TLS  *TLSConfig
}

// Server is a payment server.
type Server struct {
	conf   *Config
	logger *util.Logger
	db     *reform.DB
}

// NewServer creates a new payment server.
func NewServer(conf *Config, logger *util.Logger, db *reform.DB) *Server {
	return &Server{conf, logger, db}
}

const payPath = "/v1/pmtChannel/pay"

// ListenAndServe starts to listen and serve to requests.
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc(payPath, util.ValidateMethod(s.handlePay, http.MethodPost))

	if s.conf.TLS != nil {
		return http.ListenAndServeTLS(
			s.conf.Addr, s.conf.TLS.CertFile, s.conf.TLS.KeyFile, mux)
	}

	return http.ListenAndServe(s.conf.Addr, mux)
}
