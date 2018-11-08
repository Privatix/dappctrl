package sesssrv

import (
	"net/http"
	"time"

	"github.com/AlekSi/pointer"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/srv"
)

// StartArgs is a set of arguments for session starting.
type StartArgs struct {
	ClientID   string `json:"clientId"`
	ClientIP   string `json:"clientIp"`
	ClientPort uint16 `json:"clientPort"`
}

// StartResult is a result of session starting.
type StartResult struct {
	Offering *data.Offering `json:"offering"`
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

	var offer data.Offering
	if err := s.db.FindByPrimaryKeyTo(&offer, ch.Offering); err != nil {
		logger.Error(err.Error())
		s.RespondError(logger, w, srv.ErrInternalServerError)
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

	err := s.db.InTransaction(func(tx *reform.TX) error {
		sess := data.Session{
			ID:            util.NewUUID(),
			Channel:       ch.ID,
			Started:       now,
			LastUsageTime: now,
			ClientIP:      ip,
			ClientPort:    port,
		}
		if err := tx.Insert(&sess); err != nil {
			return err
		}

		if ch.ServiceStatus == data.ServiceActivating {
			return job.AddWithData(s.queue, tx,
				data.JobClientCompleteServiceTransition,
				data.JobChannel, ch.ID, data.JobSessionServer,
				data.ServiceActive)
		}

		return nil
	})
	if err != nil {
		s.logger.Error(err.Error())
		s.RespondError(logger, w, srv.ErrInternalServerError)
		return
	}

	s.RespondResult(logger, w, StartResult{Offering: &offer})
}
