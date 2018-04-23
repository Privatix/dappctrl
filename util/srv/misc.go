package srv

import (
	"net/http"
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
func (s *Server) RequireHTTPMethods(
	handler HandlerFunc, methods ...string) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, ctx *Context) {
		for _, v := range methods {
			if v == r.Method {
				handler(w, r, ctx)
				return
			}
		}

		s.logger.Warn("not allowed HTTP method from %s", r.RemoteAddr)
		s.RespondError(w, ErrMethodNotAllowed)
	}
}

// AuthFunc checks if a given username and password pair is correct.
type AuthFunc func(username, password string) bool

// RequireBasicAuth wraps a given handler function inside a handler with basic
// access authentication.
func (s *Server) RequireBasicAuth(
	handler HandlerFunc, auth AuthFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, ctx *Context) {
		name, pass, ok := r.BasicAuth()
		if !ok || !auth(name, pass) {
			s.logger.Warn("access denied for %s", r.RemoteAddr)
			s.RespondError(w, ErrAccessDenied)
			return
		}

		ctx.Username = name
		handler(w, r, ctx)
	}
}
