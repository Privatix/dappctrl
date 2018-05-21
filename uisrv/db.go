package uisrv

import (
	"database/sql"
	"net/http"

	reform "gopkg.in/reform.v1"
)

func (s *Server) findTo(w http.ResponseWriter, v reform.Record, id string) bool {
	if err := s.db.FindByPrimaryKeyTo(v, id); err != nil {
		if err == sql.ErrNoRows {
			s.replyNotFound(w)
			return false
		}
		s.replyUnexpectedErr(w)
		return false
	}
	return true
}

func (s *Server) insert(w http.ResponseWriter, rec reform.Struct) bool {
	if err := s.db.Insert(rec); err != nil {
		s.logger.Error("failed to insert: %v", err)
		s.replyUnexpectedErr(w)
		return false
	}
	return true
}

// Transactional funcs.

func (s *Server) begin(w http.ResponseWriter) (*reform.TX, bool) {
	tx, err := s.db.Begin()
	if err != nil {
		s.logger.Error("failed to begin transaction: %v", err)
		s.replyUnexpectedErr(w)
		return tx, false
	}
	return tx, true
}

func (s *Server) insertTx(w http.ResponseWriter, rec reform.Record, tx *reform.TX) bool {
	if err := tx.Insert(rec); err != nil {
		tx.Rollback()
		s.logger.Error("failed to insert: %v", err)
		s.replyUnexpectedErr(w)
		return false
	}
	return true
}

func (s *Server) deleteTx(w http.ResponseWriter, rec reform.Record, tx *reform.TX) bool {
	if err := tx.Delete(rec); err != nil {
		tx.Rollback()
		s.logger.Error("failed to delete: %v", err)
		s.replyUnexpectedErr(w)
		return false
	}
	return true
}

func (s *Server) commit(w http.ResponseWriter, tx *reform.TX) bool {
	if err := tx.Commit(); err != nil {
		s.logger.Warn("failed to commit transaction: %v", err)
		s.replyUnexpectedErr(w)
		return false
	}
	return true
}
