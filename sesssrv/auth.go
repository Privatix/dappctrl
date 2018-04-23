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
	var args AuthArgs
	if !s.ParseRequest(w, r, &args) {
		return
	}

	ch, ok := s.identClient(w, ctx.Username, args.ClientID)
	if !ok {
		return
	}

	if !data.ValidatePassword(args.Password, ch.Password, ch.Salt) {
		s.Logger().Warn("failed to match auth password")
		s.RespondError(w, ErrBadAuthPassword)
		return
	}

	s.RespondResult(w, nil)
}
