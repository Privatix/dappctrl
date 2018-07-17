package srv

import (
	"encoding/json"
	"net/http"
)

// Request is a server request, compatible with JSON-RPC 2.0.
type Request struct {
	JSONRPC string          `json:"jsonrpc,omitempty"`
	ID      string          `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// ParseRequest parses request handling possible errors.
func (s *Server) ParseRequest(w http.ResponseWriter,
	r *http.Request, params interface{}) (*Request, bool) {
	s.logger.Info("server request %s from %s", r.URL, r.RemoteAddr)

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Warn("failed to parse request: %s", err)
		s.RespondError(w, ErrFailedToParseRequest)
		return nil, false
	}
	r.Body.Close()

	if err := json.Unmarshal(req.Params, params); err != nil {
		s.logger.Warn("failed to parse request params: %s", err)
		s.RespondError(w, ErrFailedToParseRequest)
		return nil, false
	}

	return &req, true
}
