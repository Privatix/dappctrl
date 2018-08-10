package srv

import (
	"net/http"
)

// Server is an HTTP server.
type Server struct {
	conf *Config
	srv  http.Server
}

// TLSConfig is a TLS configuration.
type TLSConfig struct {
	CertFile string
	KeyFile  string
}

// Config is a server configuration.
type Config struct {
	Addr string
	TLS  *TLSConfig
}

// NewConfig creates a default server configuration.
func NewConfig() *Config {
	return &Config{
		Addr: "localhost:80",
		TLS:  nil,
	}
}

// NewServer creates a new HTTP server.
func NewServer(conf *Config) *Server {
	s := &Server{
		conf: conf,
		srv: http.Server{
			Addr:    conf.Addr,
			Handler: http.NewServeMux(),
		},
	}
	return s
}

// Mux is an associated http.ServeMux instance.
func (s *Server) Mux() *http.ServeMux {
	return s.srv.Handler.(*http.ServeMux)
}

// ListenAndServe starts to listen and to serve requests.
func (s *Server) ListenAndServe() error {
	if s.conf.TLS != nil {
		return s.srv.ListenAndServeTLS(
			s.conf.TLS.CertFile, s.conf.TLS.KeyFile)
	}

	return s.srv.ListenAndServe()
}

// Close immediately closes the server making ListenAndServe() to return.
func (s *Server) Close() error {
	return s.srv.Close()
}
