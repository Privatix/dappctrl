package sesssrv

import (
	"net/http"
	"time"

	"github.com/AlekSi/pointer"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/srv"
)

// StartArgs is a set of arguments for session starting.
type StartArgs struct {
	ClientID   string `json:"clientId"`
	ClientIP   string `json:"clientIp"`
	ClientPort uint16 `json:"clientPort"`
}

func (s *Server) handleStart(
	w http.ResponseWriter, r *http.Request, ctx *srv.Context) {
	logger := s.logger.Add("method", "handleStart", "sender", r.RemoteAddr)

	logger.Info("session start request from " + r.RemoteAddr)

	var args StartArgs
	if !s.ParseRequest(logger, w, r, &args) {
		return
	}

	logger = logger.Add("arguments", args)

	ch, ok := s.identClient(logger, w, ctx.Username, args.ClientID)
	if !ok {
		return
	}

	logger.Info("new client session")

	now := time.Now()

	var ip *string
	if len(args.ClientIP) != 0 {
		ip = pointer.ToString(args.ClientIP)
	}

	var port *uint16
	if args.ClientPort != 0 {
		port = pointer.ToUint16(args.ClientPort)
	}

	sess := data.Session{
		ID:            util.NewUUID(),
		Channel:       ch.ID,
		Started:       now,
		LastUsageTime: now,
		ClientIP:      ip,
		ClientPort:    port,
	}
	if err := s.db.Insert(&sess); err != nil {
		s.logger.Error(err.Error())
		s.RespondError(logger, w, srv.ErrInternalServerError)
		return
	}

	s.RespondResult(logger, w, nil)
}
