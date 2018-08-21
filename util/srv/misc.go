package srv

import (
	"net/http"

	"github.com/privatix/dappctrl/util/log"
)

// Context is a request context data.
type Context struct {
	Username string
}

// HandlerFunc is an HTTP request handler function which receives additional
// context.
type HandlerFunc func(w http.ResponseWriter, r *http.Request, ctx *Context)

// HandleFunc registers a handler function for a given pattern.
func (s *Server) HandleFunc(pattern string, handler HandlerFunc) {
	s.Mux().HandleFunc(pattern,
		func(w http.ResponseWriter, r *http.Request) {
			handler(w, r, &Context{})
		})
}

// RequireHTTPMethods wraps a given handler function inside an HTTP method
// validating handler.
func (s *Server) RequireHTTPMethods(logger log.Logger,
	handler HandlerFunc, methods ...string) HandlerFunc {
	l := logger.Add("method", "RequireHTTPMethods")
	return func(w http.ResponseWriter, r *http.Request, ctx *Context) {
		for _, v := range methods {
			if v == r.Method {
				handler(w, r, ctx)
				return
			}
		}

		l.Add("sender", r.RemoteAddr).Warn("not allowed HTTP method")
		s.RespondError(logger, w, ErrMethodNotAllowed)
	}
}

// AuthFunc checks if a given username and password pair is correct.
type AuthFunc func(username, password string) bool

// RequireBasicAuth wraps a given handler function inside a handler with basic
// access authentication.
func (s *Server) RequireBasicAuth(logger log.Logger,
	handler HandlerFunc, auth AuthFunc) HandlerFunc {
	l := logger.Add("method", "RequireBasicAuth")
	return func(w http.ResponseWriter, r *http.Request, ctx *Context) {
		name, pass, ok := r.BasicAuth()
		if !ok || !auth(name, pass) {
			l.Add("sender", r.RemoteAddr).Warn("access denied")
			s.RespondError(logger, w, ErrAccessDenied)
			return
		}

		ctx.Username = name
		handler(w, r, ctx)
	}
}
