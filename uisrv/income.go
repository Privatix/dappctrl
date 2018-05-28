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
		query = `SELECT SUM(channels.receipt_balance)
			   FROM channels
			   WHERE channels.offering=$1`
	} else if arg = r.FormValue(incomeByProduct); arg != "" {
		query = `SELECT SUM(channels.receipt_balance)
			   FROM channels
			   JOIN offerings ON offerings.product=$1
			     AND channels.offering=offerings.id`
	}

	s.replyNumFromQuery(w, query, arg)
}
