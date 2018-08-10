package srv

import (
	"encoding/json"
	"net/http"

	"github.com/privatix/dappctrl/util/log"
)

// Response is a server reply.
type Response struct {
	Error  *Error          `json:"error,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
}

func (s *Server) respond(logger log.Logger, w http.ResponseWriter, r *Response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Error != nil && r.Error.Status != 0 {
		w.WriteHeader(r.Error.Status)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(r); err != nil {
		logger.Warn("failed to send reply: " + err.Error())
	}
}

// RespondResult sends a response with a given result.
func (s *Server) RespondResult(logger log.Logger, w http.ResponseWriter,
	result interface{}) {
	logger = logger.Add("result", result)
	data, err := json.Marshal(result)
	if err != nil {
		logger.Error("failed to marhsal respond result: " + err.Error())
		s.RespondError(logger, w, ErrInternalServerError)
		return
	}

	s.respond(logger, w, &Response{Result: data})
}

// RespondError sends a response with a given error.
func (s *Server) RespondError(logger log.Logger, w http.ResponseWriter, err *Error) {
	s.respond(logger, w, &Response{Error: err})
}
