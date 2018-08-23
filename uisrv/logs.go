package uisrv

import (
	"fmt"
	"net/http"

	"github.com/privatix/dappctrl/data"
)

func (s *Server) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	var contextSearchSQL string

	searchText := r.FormValue("searchText")
	if searchText != "" {
		contextSearchSQL = fmt.Sprintf(
			"to_tsvector('english', context) @@ to_tsquery('%s:*')",
			searchText)
	}
	s.handleGetResources(w, r, &getConf{
		Params: []queryParam{
			{Name: "level", Field: "level", Op: "in"},
			{Name: "searchText", Field: "message", Op: "like"},
			{Name: "dateFrom", Field: "time", Op: ">="},
			{Name: "dateTo", Field: "time", Op: "<"},
		},
		View:      data.LogEventView,
		Paginated: true,
		FilteringSQL: filteringSQL{
			SQL:      contextSearchSQL,
			JoinWith: "OR",
		},
		OrderingSQL: "ORDER BY time DESC",
	})
}
