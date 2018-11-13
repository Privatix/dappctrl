package sesssrv

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AlekSi/pointer"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/util/srv"
)

type updateStopArgs struct {
	ClientID string `json:"clientId"`
	Units    uint64 `json:"units"`
}

func (s *Server) handleUpdateStop(logger log.Logger,
	w http.ResponseWriter, r *http.Request, ctx *srv.Context, stop bool) {

	var args updateStopArgs
	if !s.ParseRequest(logger, w, r, &args) {
		return
	}
	logger = logger.Add("arguments", args)

	ch, ok := s.identClient(logger, w, ctx.Username, args.ClientID)
	if !ok {
		return
	}

	sess, ok := s.findCurrentSession(logger, w, args.ClientID)
	if !ok {
		return
	}

	if args.Units != 0 {
		prod, ok := s.findProduct(logger, w, ctx.Username)
		if !ok {
			return
		}

		// TODO: Use unit size instead of this hardcode.
		args.Units /= 1024 * 1024

		switch prod.UsageRepType {
		case data.ProductUsageIncremental:
			sess.UnitsUsed += args.Units
		case data.ProductUsageTotal:
			sess.UnitsUsed = args.Units
		default:
			logger.Fatal("unsupported product usage")
		}
	}

	sess.LastUsageTime = time.Now()
	if stop {
		sess.Stopped = pointer.ToTime(sess.LastUsageTime)
	}

	logger = logger.Add("session", sess)
	logger.Info("updating session")

	err := s.db.InTransaction(func(tx *reform.TX) error {
		if err := tx.Save(sess); err != nil {
			return err
		}

		if stop {
			var status string
			if ch.ServiceStatus == data.ServiceTerminating {
				status = data.ServiceTerminated
			} else {
				status = data.ServiceSuspended
			}

			err := job.AddWithData(s.queue, tx,
				data.JobClientCompleteServiceTransition,
				data.JobChannel, ch.ID, data.JobSessionServer,
				status)
			if err != nil && err != job.ErrDuplicatedJob {
				return err
			}
		}

		return nil
	})
	if err != nil {
		logger.Error(err.Error())
		s.RespondError(logger, w, srv.ErrInternalServerError)
		return
	}

	s.RespondResult(logger, w, nil)
}

// UpdateArgs is a set of arguments for session usage update.
type UpdateArgs = updateStopArgs

func (s *Server) handleUpdate(
	w http.ResponseWriter, r *http.Request, ctx *srv.Context) {
	logger := s.logger.Add("method", "handleUpdate", "sender", r.RemoteAddr)

	logger.Info("session update from " + r.RemoteAddr)

	s.handleUpdateStop(logger, w, r, ctx, false)
}

// StopArgs is a set of arguments for session stopping.
type StopArgs = updateStopArgs

func (s *Server) handleStop(
	w http.ResponseWriter, r *http.Request, ctx *srv.Context) {
	logger := s.logger.Add("method", "handleStop", "url", r.URL, "sender",
		r.RemoteAddr)

	logger.Info("session stop from " + r.RemoteAddr)

	s.handleUpdateStop(logger, w, r, ctx, true)
}
