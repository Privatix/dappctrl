package srv

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"golang.org/x/crypto/sha3"

	"github.com/privatix/dappctrl/data"
)

// AuthRequest is an authentication request.
type AuthRequest struct {
	Channel  string `json:"channel"`
	Password string `json:"password"`
}

// AuthReply is an authentication reply.
type AuthReply struct {
	errorReply
}

func checkPassword(ch *data.Channel, pwd string) bool {
	hash := sha3.Sum256([]byte(pwd + fmt.Sprint(ch.Salt)))
	return base64.URLEncoding.EncodeToString(hash[:]) == ch.Password
}

func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if !s.parseRequest(w, r, &req) {
		return
	}

	s.logger.Info("channel: %s", req.Channel)

	ch := data.Channel{ID: req.Channel}
	if !s.findByPrimaryKey(w, &ch, false) {
		return
	}

	if !checkPassword(&ch, req.Password) {
		s.logger.Warn("channel password mismatch")
		w.WriteHeader(http.StatusForbidden)
		s.reply(w, errorReply{ErrAccessDenied})
		return
	}

	s.reply(w, AuthReply{})
}
