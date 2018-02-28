package vpn

import (
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	vpnutil "github.com/privatix/dappctrl/vpn/util"
	"net"
	"net/http"
	"time"
)

// StartRequest is a request to start a client session.
type StartRequest struct {
	Channel    string `json:"channel"`
	ServerIP   string `json:"serverIp"`
	ClientIP   string `json:"clientIp"`
	ClientPort uint16 `json:"clientPort"`
}

// StartReply is a reply for a session start request.
type StartReply struct {
	errorReply
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	var req StartRequest
	if !s.parseRequest(w, r, &req) {
		return
	}

	s.logger.Info("channel: %s", req.Channel)
	s.logger.Info("peers: %s:%d -> %s",
		req.ClientIP, req.ClientPort, req.ServerIP)

	sip := net.ParseIP(req.ServerIP)
	cip := net.ParseIP(req.ClientIP)

	if sip == nil || sip.IsUnspecified() ||
		cip == nil || cip.IsUnspecified() || req.ClientPort == 0 {
		s.logger.Warn("malformed request")
		s.reply(w, errorReply{ErrMalformedRequest})
		return
	}

	ch := data.Channel{ID: req.Channel}
	if !s.findByPrimaryKey(w, &ch, false) {
		return
	}

	positive, err := vpnutil.HasPositiveBalance(s.db, req.Channel)
	if err != nil {
		s.logger.Error("failed to check for positive balance: %s", err)
		s.replyInternalError(w)
		return
	}

	if !positive {
		s.logger.Warn("no positive balance for channel %s", req.Channel)
		s.reply(w, errorReply{ErrNoPositiveBalance})
		return
	}

	tx, ok := s.begin(w)
	if !ok {
		return
	}

	sess := data.Session{
		ID:      util.NewUUID(),
		Channel: req.Channel,
		Started: time.Now(),
	}

	if !s.insert(w, tx, &sess) {
		return
	}

	vsess := data.VPNSession{
		ID:         sess.ID,
		ServerIP:   req.ServerIP,
		ClientIP:   req.ClientIP,
		ClientPort: req.ClientPort,
	}

	if !s.insert(w, tx, &vsess) {
		return
	}

	if !s.commit(w, tx) {
		return
	}

	s.reply(w, StartReply{})
}
