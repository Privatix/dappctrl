package sesssrv

import (
	"net/http"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/srv"
)

// AuthArgs is a set of authentication arguments.
type AuthArgs struct {
	ClientID string `json:"clientId"`
	Password string `json:"password"`
}

func (s *Server) handleAuth(
	w http.ResponseWriter, r *http.Request, ctx *srv.Context) {
	logger := s.logger.Add("method", "handleAuth", "sender", r.RemoteAddr)

	logger.Info("session auth request from " + r.RemoteAddr)

	var args AuthArgs
	if !s.ParseRequest(logger, w, r, &args) {
		return
	}
	logger = logger.Add("arguments", args)

	ch, ok := s.identClient(logger, w, ctx.Username, args.ClientID)
	if !ok {
		return
	}

	if data.ValidatePassword(ch.Password, args.Password,
		string(ch.Salt)) != nil {
		logger.Warn("failed to match auth password")
		s.RespondError(logger, w, ErrBadAuthPassword)
		return
	}

	s.RespondResult(logger, w, nil)
}
