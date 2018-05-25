package uisrv

import (
	"net/http"
)

// Params to compute income by.
const (
	incomeByOffering = "offering"
	incomeByProduct  = "product"
)

func (s *Server) handleGetIncome(w http.ResponseWriter, r *http.Request) {
	var arg, query string

	if arg = r.FormValue(incomeByOffering); arg != "" {
		query = `select sum(receipt_balance)
			   from channels
			   where channels.offering=$1`
	} else if arg = r.FormValue(incomeByProduct); arg != "" {
		query = `select sum(receipt_balance)
			   from channels
			   join offerings on offerings.product=$1
			     and channels.offering=offerings.id`
	}

	s.replyNumFromQuery(w, query, arg)
}
