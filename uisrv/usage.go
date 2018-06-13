package uisrv

import (
	"net/http"
)

// Params to compute usage by.
const (
	usagesByChannelID  = "channel"
	usagesByOfferingID = "offering"
	usagesByProductID  = "product"
)

func (s *Server) handleGetUsage(w http.ResponseWriter, r *http.Request) {
	var query, arg string

	arg = r.FormValue(usagesByChannelID)
	if arg != "" {
		query = `select sum(sessions.units_used)
			   from sessions where channel=$1`
	} else if arg = r.FormValue(usagesByOfferingID); arg != "" {
		query = `select sum(sessions.units_used)
			   from sessions
			   join channels on channels.offering=$1`
	} else if arg = r.FormValue(usagesByProductID); arg != "" {
		query = `select sum(sessions.units_used)
			   from sessions
			   join offerings on offerings.product=$1
			   join channels on channels.offering=offerings.id`
	}

	s.replyNumFromQuery(w, query, arg)
}
