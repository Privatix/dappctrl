package uisrv

import "net/http"

func (s *Server) pageNotFound(w http.ResponseWriter, r *http.Request) {
	s.logger.Warn("page not found at: %s", r.URL.Path)
	w.WriteHeader(http.StatusNotFound)
}
