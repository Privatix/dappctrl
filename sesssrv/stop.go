package sesssrv

import (
	"net/http"
	"time"

	"github.com/AlekSi/pointer"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/srv"
)

type updateStopArgs struct {
	ClientID string `json:"clientId"`
	Units    uint64 `json:"units"`
}

func (s *Server) handleUpdateStop(
	w http.ResponseWriter, r *http.Request, ctx *srv.Context, stop bool) {
	var args updateStopArgs
	if !s.ParseRequest(w, r, &args) {
		return
	}

	_, ok := s.identClient(w, ctx.Username, args.ClientID)
	if !ok {
		return
	}

	sess, ok := s.findCurrentSession(w, args.ClientID)
	if !ok {
		return
	}

	if args.Units != 0 {
		prod, ok := s.findProduct(w, ctx.Username)
		if !ok {
			return
		}

		switch prod.UsageRepType {
		case data.ProductUsageIncremental:
			sess.UnitsUsed += args.Units
		case data.ProductUsageTotal:
			sess.UnitsUsed = args.Units
		default:
			panic("unsupported product usage: " + prod.UsageRepType)
		}
	}

	sess.LastUsageTime = time.Now()
	if stop {
		sess.Stopped = pointer.ToTime(sess.LastUsageTime)
	}

	if err := s.db.Save(sess); err != nil {
		s.Logger().Error("failed to save session: %s", err)
		s.RespondError(w, srv.ErrInternalServerError)
	}

	s.RespondResult(w, nil)
}

// UpdateArgs is a set of arguments for session usage update.
type UpdateArgs = updateStopArgs

func (s *Server) handleUpdate(
	w http.ResponseWriter, r *http.Request, ctx *srv.Context) {
	s.handleUpdateStop(w, r, ctx, false)
}

// StopArgs is a set of arguments for session stopping.
type StopArgs = updateStopArgs

func (s *Server) handleStop(
	w http.ResponseWriter, r *http.Request, ctx *srv.Context) {
	s.handleUpdateStop(w, r, ctx, true)
}
