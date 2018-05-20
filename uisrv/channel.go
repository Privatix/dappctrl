package uisrv

import (
	"net/http"

	"github.com/privatix/dappctrl/data"
)

var channelsGetParams = []queryParam{
	{Name: "id", Field: "id"},
	{Name: "serviceStatus", Field: "service_status"},
}

// handleChannels calls appropriate handler by scanning incoming request.
func (s *Server) handleChannels(w http.ResponseWriter, r *http.Request) {
	if id := idFromStatusPath(channelsPath, r.URL.Path); id != "" {
		if r.Method == "GET" {
			s.handleGetChannelStatus(w, r, id)
			return
		}
		if r.Method == "PUT" {
			s.handlePutChannelStatus(w, r, id)
		}
	} else {
		if r.Method == "GET" {
			s.handleGetChannels(w, r)
			return
		}
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleGetChannels replies with all channels or a channel by id
// available to the agent.
func (s *Server) handleGetChannels(w http.ResponseWriter, r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params: channelsGetParams,

		View:         data.ChannelTable,
		FilteringSQL: `channels.agent IN (SELECT eth_addr FROM accounts)`,
	})
}

// handleGetChannels replies with all channels or a channel by id
// available to the client.
func (s *Server) handleGetClientChannels(w http.ResponseWriter,
	r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params: channelsGetParams,

		View:         data.ChannelTable,
		FilteringSQL: `channels.agent NOT IN (SELECT eth_addr FROM accounts)`,
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
