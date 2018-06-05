package uisrv

import (
	"net/http"

	"github.com/privatix/dappctrl/data"
)

func (s *Server) handleTransactions(w http.ResponseWriter, r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params: nil,
		View:   data.EthTxTable,
	})
}
