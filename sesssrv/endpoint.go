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
	logger := s.logger.Add("method", "handleEndpointMsg", "sender", r.RemoteAddr)

	logger.Info("session endpoint msg request from " + r.RemoteAddr)

	var args EndpointMsgArgs
	if !s.ParseRequest(logger, w, r, &args) {
		return
	}
	logger = logger.Add("arguments", args)

	if args.ChannelID == "" {
		s.RespondError(logger, w, ErrEndpointNotFound)
		return
	}

	ept, ok := s.findEndpoint(logger, w, args.ChannelID)
	if !ok {
		return
	}

	s.RespondResult(logger, w, ept)
}
