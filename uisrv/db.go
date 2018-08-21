package uisrv

import (
	"database/sql"
	"fmt"
	"net/http"

	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/util/log"
)

func (s *Server) findTo(logger log.Logger, w http.ResponseWriter, v reform.Record, id string) bool {
	if err := s.db.FindByPrimaryKeyTo(v, id); err != nil {
		if err == sql.ErrNoRows {
			s.replyNotFound(logger, w)
			return false
		}
		logger.Error(fmt.Sprintf("failed to find %T: %v", v, err))
		s.replyUnexpectedErr(logger, w)
		return false
	}
	return true
}

func (s *Server) selectOneTo(logger log.Logger, w http.ResponseWriter, v reform.Record,
	filter string, args ...interface{}) bool {
	if err := s.db.SelectOneTo(v, filter, args...); err != nil {
		if err == sql.ErrNoRows {
			s.replyNotFound(logger, w)
			return false
		}
		logger.Error(fmt.Sprintf("failed to find %T: %v", v, err))
		s.replyUnexpectedErr(logger, w)
		return false
	}
	return true
}

func (s *Server) insert(logger log.Logger, w http.ResponseWriter, rec reform.Struct) bool {
	if err := s.db.Insert(rec); err != nil {
		logger.Error(fmt.Sprintf("failed to insert: %v", err))
		s.replyUnexpectedErr(logger, w)
		return false
	}
	return true
}

// Transactional funcs.

func (s *Server) begin(logger log.Logger, w http.ResponseWriter) (*reform.TX, bool) {
	tx, err := s.db.Begin()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to begin transaction: %v", err))
		s.replyUnexpectedErr(logger, w)
		return tx, false
	}
	return tx, true
}

func (s *Server) insertTx(logger log.Logger, w http.ResponseWriter,
	rec reform.Record, tx *reform.TX) bool {
	if err := tx.Insert(rec); err != nil {
		tx.Rollback()
		logger.Error(fmt.Sprintf("failed to insert: %v", err))
		s.replyUnexpectedErr(logger, w)
		return false
	}
	return true
}

func (s *Server) updateTx(logger log.Logger, w http.ResponseWriter,
	rec reform.Record, tx *reform.TX) bool {
	if err := tx.Update(rec); err != nil {
		tx.Rollback()
		logger.Error(fmt.Sprintf("failed to update: %v", err))
		s.replyUnexpectedErr(logger, w)
		return false
	}
	return true
}

func (s *Server) deleteTx(logger log.Logger,
	w http.ResponseWriter, rec reform.Record, tx *reform.TX) bool {
	if err := tx.Delete(rec); err != nil {
		tx.Rollback()
		logger.Error(fmt.Sprintf("failed to delete: %v", err))
		s.replyUnexpectedErr(logger, w)
		return false
	}
	return true
}

func (s *Server) commit(logger log.Logger,
	w http.ResponseWriter, tx *reform.TX) bool {
	if err := tx.Commit(); err != nil {
		logger.Warn(fmt.Sprintf("failed to commit transaction: %v", err))
		s.replyUnexpectedErr(logger, w)
		return false
	}
	return true
}
