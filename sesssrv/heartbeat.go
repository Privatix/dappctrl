package sesssrv

import (
	"net/http"

	"github.com/privatix/dappctrl/util/srv"
)

// HeartbeatArgs is a set of heartbeat arguments.
type HeartbeatArgs struct {
	ClientID string `json:"clientId"`
}

// HeartbeatResult is a result of heartbeat request.
type HeartbeatResult struct {
	ServiceStatus string `json:"serviceStatus"`
}

func (s *Server) handleHeartbeat(
	w http.ResponseWriter, r *http.Request, ctx *srv.Context) {
	logger := s.logger.Add(
		"method", "handleHeartbeat", "sender", r.RemoteAddr)

	var args HeartbeatArgs
	if !s.ParseRequest(logger, w, r, &args) {
		return
	}

	logger = logger.Add("arguments", args)

	ch, ok := s.findChannel(logger, w, args.ClientID)
	if !ok {
		return
	}

	s.RespondResult(logger, w,
		HeartbeatResult{ServiceStatus: ch.ServiceStatus})
}
