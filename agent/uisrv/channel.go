package uisrv

import (
	"net/http"

	"github.com/privatix/dappctrl/data"
)

// handleChannels calls appropriate handler by scanning incoming request.
func (s *Server) handleChannels(w http.ResponseWriter, r *http.Request) {
	if id := idFromStatusPath(channelsPath, r.URL.Path); id != "" {
		if r.Method == "GET" {
			basicAuthMiddlewareFunc(s, func(w http.ResponseWriter, r *http.Request) {
				s.handleGetChannelStatus(w, r, id)
			})(w, r)
			return
		}
		if r.Method == "PUT" {
			basicAuthMiddlewareFunc(s, func(w http.ResponseWriter, r *http.Request) {
				s.handlePutChannelStatus(w, r, id)
			})(w, r)
		}
	} else {
		if r.Method == "GET" {
			basicAuthMiddlewareFunc(s, s.handleGetChannels)(w, r)
			return
		}
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleGetChannels replies with all channels or a channel by id.
func (s *Server) handleGetChannels(w http.ResponseWriter, r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params: []queryParam{{Name: "id", Field: "id"}},
		View:   data.ChannelTable,
	})
}

// handleGetChannelStatus replies with channels status by id.
func (s *Server) handleGetChannelStatus(w http.ResponseWriter, r *http.Request, id string) {
	channel := &data.Channel{}
	if !s.findTo(w, channel, id) {
		return
	}
	s.replyStatus(w, channel.ChannelStatus)
}

func (s *Server) handlePutChannelStatus(w http.ResponseWriter, r *http.Request, id string) {
	payload := &ActionPayload{}
	if !s.parsePayload(w, r, payload) {
		return
	}
	s.logger.Info("action ( %v )  request for channel with id: %v recieved.", payload.Action, id)
	// TODO once job queue implemented.
}
