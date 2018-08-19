package uisrv

import (
	"fmt"
	"net/http"
)

func (s *Server) pageNotFound(w http.ResponseWriter, r *http.Request) {
	s.logger.Add("method", "pageNotFound").Warn(
		fmt.Sprintf("page not found at: %s", r.URL.Path))
	w.WriteHeader(http.StatusNotFound)
}
