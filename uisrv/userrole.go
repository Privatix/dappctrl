package uisrv

import "net/http"

func (s *Server) handleGetUserRole(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.Add("method", "handleGetUserRole")
	s.reply(logger, w, s.dappRole)
}
