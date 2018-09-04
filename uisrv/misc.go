package uisrv

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
	validator "gopkg.in/go-playground/validator.v9"
)

const timeFormat = "2006-01-02T15:04:05.999Z"

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

func (s *Server) parsePayload(logger log.Logger, w http.ResponseWriter,
	r *http.Request, v interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		logger.Warn(fmt.Sprintf("failed to parse request body: %v", err))
		s.replyInvalidRequest(logger, w)
		return false
	}
	return true
}

func (s *Server) replyErr(logger log.Logger,
	w http.ResponseWriter, status int, reply *serverError) {
	w.WriteHeader(status)
	s.reply(logger, w, reply)
}

func (s *Server) replyNotFound(logger log.Logger, w http.ResponseWriter) {
	s.replyErr(logger, w, http.StatusNotFound, &serverError{
		Message: "requested resources was not found",
	})
}

type replyOK struct {
	Message string `json:"message"`
}

func (s *Server) replyOK(logger log.Logger, w http.ResponseWriter, msg string) {
	s.reply(logger, w, &replyOK{msg})
}

func (s *Server) replyUnexpectedErr(logger log.Logger, w http.ResponseWriter) {
	s.replyErr(logger, w, http.StatusInternalServerError, &serverError{
		Message: "An unexpected error occurred",
	})
}

func (s *Server) replyInvalidRequest(logger log.Logger, w http.ResponseWriter) {
	s.replyErr(logger, w, http.StatusBadRequest, &serverError{
		Message: "",
	})
}

func (s *Server) replyInvalidAction(logger log.Logger, w http.ResponseWriter) {
	s.replyErr(logger, w, http.StatusBadRequest, &serverError{
		Message: "invalid action",
	})
}

type replyEntity struct {
	ID interface{} `json:"id"`
}

func (s *Server) replyEntityCreated(
	logger log.Logger, w http.ResponseWriter, id interface{}) {
	w.WriteHeader(http.StatusCreated)
	s.reply(logger, w, &replyEntity{ID: id})
}

func (s *Server) replyEntityUpdated(
	logger log.Logger, w http.ResponseWriter, id interface{}) {
	s.reply(logger, w, &replyEntity{ID: id})
}

type statusReply struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
}

func (s *Server) replyStatus(
	logger log.Logger, w http.ResponseWriter, status string) {
	s.reply(logger, w, &statusReply{Status: status})
}

func (s *Server) reply(logger log.Logger, w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Warn(fmt.Sprintf("failed to marshal: %v", err))
	}
}

func (s *Server) replyNumFromQuery(logger log.Logger, w http.ResponseWriter, query, arg string) {
	row := s.db.QueryRow(query, arg)
	var queryRet sql.NullInt64
	if err := row.Scan(&queryRet); err != nil {
		logger.Error(fmt.Sprintf("failed to get usage: %v", err))
		s.replyUnexpectedErr(logger, w)
		return
	}

	retB, err := json.Marshal(&queryRet.Int64)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to encode usage: %v", err))
		s.replyUnexpectedErr(logger, w)
		return
	}
	w.Write(retB)
	return
}

func (s *Server) defaultGasPrice(logger log.Logger, w http.ResponseWriter) (uint64, bool) {
	gasPriceSettings := &data.Setting{}
	if err := data.FindOneTo(s.db.Querier, gasPriceSettings,
		"key", data.SettingDefaultGasPrice); err != nil {
		logger.Error(err.Error())
		s.replyUnexpectedErr(logger, w)
		return 0, false
	}

	val, err := strconv.ParseUint(gasPriceSettings.Value, 10, 64)
	if err != nil {
		logger.Error(
			fmt.Sprintf("failed to parse default gas price: %v", err))
		s.replyUnexpectedErr(logger, w)
		return 0, false
	}

	return val, true
}

func singleTimeFormat(tm time.Time) string {
	return tm.Format(timeFormat)
}

func singleTimeFormatFromStr(ti string) string {
	tm, err := time.Parse(time.RFC3339Nano, ti)
	if err != nil {
		return ""
	}
	return singleTimeFormat(tm)
}
