package srv

import (
	"encoding/json"
	"net/http"
)

// Request is a server request.
type Request struct {
	Args json.RawMessage `json:"args,omitempty"`
}

// ParseRequest parses request handling possible errors.
func (s *Server) ParseRequest(
	w http.ResponseWriter, r *http.Request, args interface{}) bool {
	s.logger.Info("server request %s from %s", r.URL, r.RemoteAddr)

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Warn("failed to parse request: %s", err)
		s.RespondError(w, ErrFailedToParseRequest)
		return false
	}
	r.Body.Close()

	if err := json.Unmarshal(req.Args, args); err != nil {
		s.logger.Warn("failed to parse request arguments: %s", err)
		s.RespondError(w, ErrFailedToParseRequest)
		return false
	}

	return true
}
