package srv

import (
	"encoding/json"
	"net/http"
)

// Response is a server reply.
type Response struct {
	Error  *Error          `json:"error,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
}

func (s *Server) respond(w http.ResponseWriter, r *Response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Error != nil && r.Error.Status != 0 {
		w.WriteHeader(r.Error.Status)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(r); err != nil {
		s.logger.Warn("failed to send reply: %s", err)
	}
}

// RespondResult sends a response with a given result.
func (s *Server) RespondResult(w http.ResponseWriter, result interface{}) {
	data, err := json.Marshal(result)
	if err != nil {
		s.logger.Error("failed to marhsal respond result: %s", err)
		s.RespondError(w, ErrInternalServerError)
		return
	}

	s.respond(w, &Response{Result: data})
}

// RespondError sends a response with a given error.
func (s *Server) RespondError(w http.ResponseWriter, err *Error) {
	s.respond(w, &Response{Error: err})
}
