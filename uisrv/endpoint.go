package uisrv

import (
	"net/http"

	"github.com/privatix/dappctrl/data"
)

// handleGetEndpoints calls appropriate handler by scanning incoming request.
func (s *Server) handleGetEndpoints(w http.ResponseWriter, r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params: []queryParam{
			{Name: "ch_id", Field: "channel"},
			{Name: "id", Field: "id"},
			{Name: "template", Field: "template"},
		},
		View: data.EndpointUITable,
	})
}
