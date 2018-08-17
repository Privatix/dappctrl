package uisrv

import (
	"net/http"

	"github.com/privatix/dappctrl/data"
)

func (s *Server) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params: []queryParam{
			{Name: "level", Field: "level", Op: "in"},
			{Name: "dateFrom", Field: "time", Op: ">="},
			{Name: "dateTo", Field: "time", Op: "<"},
		},
		View:      data.LogEventView,
		Paginated: true,
	})
}
