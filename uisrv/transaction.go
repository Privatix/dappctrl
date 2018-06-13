package uisrv

import (
	"net/http"

	"github.com/privatix/dappctrl/data"
)

func (s *Server) handleTransactions(w http.ResponseWriter, r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params: []queryParam{
			{Name: "relatedType", Field: "related_type", Op: "="},
			{Name: "relatedID", Field: "related_id", Op: "="},
		},
		View: data.EthTxTable,
	})
}
