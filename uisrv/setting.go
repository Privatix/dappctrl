package uisrv

import (
	"net/http"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
)

// handleSettings calls appropriate handler by scanning incoming request.
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		s.handleGetSettings(w, r)
		return
	}
	if r.Method == "PUT" {
		s.handlePutSettings(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleGetSettings replies with settings.
func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params: []queryParam{
			{Name: "key", Field: "key"},
		},
		View: data.SettingTable,
		FilteringSQL: filteringSQL{
			SQL:      "permissions > 0",
			JoinWith: "AND",
		},
	})
}

type settingPayload []data.Setting

// handlePutSettings updates settings.
func (s *Server) handlePutSettings(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.Add("method", "handlePutSettings")

	var payload settingPayload
	if !s.parsePayload(logger, w, r, &payload) {
		return
	}

	tx, ok := s.begin(logger, w)
	if !ok {
		return
	}
	defer tx.Rollback()

	for _, setting := range payload {
		var settingFromDB data.Setting

		// gets setting from database
		if err := tx.FindByPrimaryKeyTo(&settingFromDB,
			setting.Key); err != nil {
			if err != reform.ErrNoRows {
				s.replyUnexpectedErr(logger, w)
				return
			}
		}

		// if setting exists and settings.permissions != 2
		// then setting ignored
		if settingFromDB.Key != "" {
			if settingFromDB.Permissions != data.ReadWrite {
				continue
			}

			// copy permissions from database
			setting.Permissions = settingFromDB.Permissions
		}

		if !s.updateTx(logger, w, &setting, tx) {
			return
		}
	}

	if !s.commit(logger, w, tx) {
		return
	}

	s.replyOK(logger, w, "updated.")
}
