package uisrv

import (
	"net/http"

	"github.com/privatix/dappctrl/data"
)

// handleSessions returns all sessions or sessions by channel id.
func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params: []queryParam{{Name: "channelId", Field: "channel"}},
		View:   data.SessionTable,
	})
}
