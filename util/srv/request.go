package srv

import (
	"encoding/json"
	"net/http"

	"github.com/privatix/dappctrl/util/log"
)

// Request is a server request.
type Request struct {
	Args json.RawMessage `json:"args,omitempty"`
}

// ParseRequest parses request handling possible errors.
func (s *Server) ParseRequest(logger log.Logger,
	w http.ResponseWriter, r *http.Request, args interface{}) bool {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("failed to parse request: " + err.Error())
		s.RespondError(logger, w, ErrFailedToParseRequest)
		return false
	}
	r.Body.Close()

	if err := json.Unmarshal(req.Args, args); err != nil {
		logger.Add("arguments", req.Args).Warn(
			"failed to parse request arguments: " + err.Error())
		s.RespondError(logger, w, ErrFailedToParseRequest)
		return false
	}

	return true
}
