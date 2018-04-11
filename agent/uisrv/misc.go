package uisrv

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"gopkg.in/reform.v1"

	validator "gopkg.in/go-playground/validator.v9"

	"github.com/privatix/dappctrl/data"
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

// idFromStatusPath returns id from path of format prefix/id/status.
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

func invalidUnitType(v string) bool {
	return v != data.UnitScalar && v != data.UnitSeconds
}

func invalidBillingType(v string) bool {
	return v != data.BillingPrepaid && v != data.BillingPostpaid
}

func (s *Server) parseOfferingPayload(w http.ResponseWriter,
	r *http.Request, offering *data.Offering) bool {
	if !s.parsePayload(w, r, offering) ||
		validate.Struct(offering) != nil ||
		invalidUnitType(offering.UnitType) ||
		invalidBillingType(offering.BillingType) {
		s.replyInvalidPayload(w)
		return false
	}
	return true
}

func (s *Server) parseProductPayload(w http.ResponseWriter,
	r *http.Request, product *data.Product) bool {
	if !s.parsePayload(w, r, product) ||
		validate.Struct(product) != nil ||
		product.OfferTplID == nil ||
		product.OfferAccessID == nil ||
		(product.UsageRepType != data.ProductUsageIncremental &&
			product.UsageRepType != data.ProductUsageTotal) {
		s.replyInvalidPayload(w)
		return false
	}
	return true
}

func invalidTemplateKind(v string) bool {
	return v != data.TemplateOffer && v != data.TemplateAccess
}

func (s *Server) parseTemplatePayload(w http.ResponseWriter,
	r *http.Request, tpl *data.Template) bool {
	v := make(map[string]interface{})
	if !s.parsePayload(w, r, tpl) ||
		invalidTemplateKind(tpl.Kind) ||
		json.Unmarshal(tpl.Raw, &v) != nil {
		s.replyInvalidPayload(w)
		return false
	}
	return true
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

func (s *Server) findByID(w http.ResponseWriter, v reform.Record, id string) bool {
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

func (s *Server) replyErr(w http.ResponseWriter, reply *serverError) {
	w.WriteHeader(reply.Code)
	s.reply(w, reply)
}

func (s *Server) replyNotFound(w http.ResponseWriter) {
	s.replyErr(w, &serverError{
		Code:    http.StatusNotFound,
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
	s.replyErr(w, &serverError{
		Code:    http.StatusInternalServerError,
		Message: "An unexpected error occurred",
	})
}

func (s *Server) replyInvalidPayload(w http.ResponseWriter) {
	s.replyErr(w, &serverError{
		Code:    http.StatusBadRequest,
		Message: "",
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
	s.reply(w, &statusReply{Code: http.StatusOK, Status: status})
}

func (s *Server) reply(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.logger.Warn("failed to marshal: %v", err)
	}
}
