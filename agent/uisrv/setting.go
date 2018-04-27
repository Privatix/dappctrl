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

// handleGetSettings replies with all settings.
func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params: nil,
		View:   data.SettingTable,
		Transform: func(rec reform.Struct) interface{} {
			setting, ok := rec.(*data.Setting)
			if !ok || setting.Key == passwordKey || setting.Key == saltKey {
				return nil
			}
			return setting
		},
	})
}

type settingPayload []data.Setting

// handlePutSettings updates settings.
func (s *Server) handlePutSettings(w http.ResponseWriter, r *http.Request) {
	var payload settingPayload
	if !s.parsePayload(w, r, &payload) {
		return
	}
	tx, err := s.db.Begin()
	if err != nil {
		s.logger.Error("failed to update settings: %v", err)
		s.replyUnexpectedErr(w)
		return
	}
	for _, setting := range payload {
		err := tx.Update(&setting)
		if err != nil {
			tx.Rollback()
			s.replyUnexpectedErr(w)
			return
		}
	}
	if tx.Commit() != nil {
		s.replyUnexpectedErr(w)
		return
	}
	s.replyOK(w, "updated.")
}
