package uisrv

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	validator "gopkg.in/go-playground/validator.v9"
)

var (
	validate = validator.New()
)

// serverError is a server reply on unexpected error.
type serverError struct {
	// Code is a status code.
	Code int `json:"code"`
	// Message is a description of the error.
	Message string `json:"message"`
}

// idFromStatusPath returns id from path of format {prefix}{id}/status.
func idFromStatusPath(prefix, path string) string {
	parts := strings.Split(path, prefix)
	if len(parts) != 2 {
		return ""
	}
	parts = strings.Split(parts[1], "/")
	if len(parts) != 2 || parts[1] != "status" {
		return ""
	}
	return parts[0]
}

func (s *Server) parsePayload(w http.ResponseWriter,
	r *http.Request, v interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		s.logger.Warn("failed to parse request body: %v", err)
		s.replyInvalidPayload(w)
		return false
	}
	return true
}

func (s *Server) replyErr(w http.ResponseWriter, status int, reply *serverError) {
	w.WriteHeader(status)
	s.reply(w, reply)
}

func (s *Server) replyNotFound(w http.ResponseWriter) {
	s.replyErr(w, http.StatusNotFound, &serverError{
		Message: "requested resources was not found",
	})
}

type replyOK struct {
	Message string `json:"message"`
}

func (s *Server) replyOK(w http.ResponseWriter, msg string) {
	s.reply(w, &replyOK{msg})
}

func (s *Server) replyUnexpectedErr(w http.ResponseWriter) {
	s.replyErr(w, http.StatusInternalServerError, &serverError{
		Message: "An unexpected error occurred",
	})
}

func (s *Server) replyInvalidPayload(w http.ResponseWriter) {
	s.replyErr(w, http.StatusBadRequest, &serverError{
		Message: "",
	})
}

func (s *Server) replyInvalidAction(w http.ResponseWriter) {
	s.replyErr(w, http.StatusBadRequest, &serverError{
		Message: "invalid action",
	})
}

type replyEntity struct {
	ID interface{} `json:"id"`
}

func (s *Server) replyEntityCreated(w http.ResponseWriter, id interface{}) {
	w.WriteHeader(http.StatusCreated)
	s.reply(w, &replyEntity{ID: id})
}

func (s *Server) replyEntityUpdated(w http.ResponseWriter, id interface{}) {
	s.reply(w, &replyEntity{ID: id})
}

type statusReply struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
}

func (s *Server) replyStatus(w http.ResponseWriter, status string) {
	s.reply(w, &statusReply{Status: status})
}

func (s *Server) reply(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.logger.Warn("failed to marshal: %v", err)
	}
}

func (s *Server) replyNumFromQuery(w http.ResponseWriter, query, arg string) {
	row := s.db.QueryRow(query, arg)
	var queryRet sql.NullInt64
	if err := row.Scan(&queryRet); err != nil {
		s.logger.Error("failed to get usage: %v", err)
		s.replyUnexpectedErr(w)
		return
	}

	retB, err := json.Marshal(&queryRet.Int64)
	if err != nil {
		s.logger.Error("failed to encode usage: %v", err)
		s.replyUnexpectedErr(w)
		return
	}
	w.Write(retB)
	return
}
