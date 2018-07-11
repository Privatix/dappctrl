package sesssrv

import (
	"net/http"

	"github.com/privatix/dappctrl/util/srv"
)

// EndpointMsgArgs is a set of endpoint message arguments.
type EndpointMsgArgs struct {
	ChannelID string `json:"channel"`
}

func (s *Server) handleEndpointMsg(
	w http.ResponseWriter, r *http.Request, ctx *srv.Context) {
	var args EndpointMsgArgs
	if !s.ParseRequest(w, r, &args) {
		return
	}

	if args.ChannelID == "" {
		s.RespondError(w, ErrEndpointNotFound)
		return
	}

	ept, ok := s.findEndpoint(w, args.ChannelID)
	if !ok {
		return
	}

	s.RespondResult(w, ept)
}
