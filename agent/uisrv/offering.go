package uisrv

import (
	"net/http"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

// handleOfferings calls appropriate handler by scanning incoming request.
func (s *Server) handleOfferings(w http.ResponseWriter, r *http.Request) {
	if id := idFromStatusPath(offeringsPath, r.URL.Path); id != "" {
		if r.Method == "PUT" {
			s.handlePutOfferingStatus(w, r, id)
			return
		}
		if r.Method == "GET" {
			s.handleGetOfferingStatus(w, r, id)
			return
		}
	} else {
		if r.Method == "POST" {
			s.handlePostOffering(w, r)
			return
		}
		if r.Method == "PUT" {
			s.handlePutOffering(w, r)
			return
		}
		if r.Method == "GET" {
			s.handleGetOfferings(w, r)
			return
		}
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handlePostOffering creates offering.
func (s *Server) handlePostOffering(w http.ResponseWriter, r *http.Request) {
	offering := &data.Offering{}
	if !s.parseOfferingPayload(w, r, offering) {
		return
	}
	err := s.fillOffering(offering)
	if err != nil {
		s.replyInvalidPayload(w)
		return
	}
	if !s.insert(w, offering) {
		return
	}
	s.replyEntityCreated(w, offering.ID)
}

// handlePutOffering updates offering.
func (s *Server) handlePutOffering(w http.ResponseWriter, r *http.Request) {
	offering := &data.Offering{}
	if !s.parseOfferingPayload(w, r, offering) {
		return
	}
	err := s.fillOffering(offering)
	if err != nil {
		s.replyUnexpectedErr(w)
		return
	}
	if err := s.db.Update(offering); err != nil {
		s.logger.Warn("failed to update offering: %v", err)
		s.replyUnexpectedErr(w)
		return
	}
	s.replyEntityUpdated(w, offering.ID)
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

func invalidUnitType(v string) bool {
	return v != data.UnitScalar && v != data.UnitSeconds
}

func invalidBillingType(v string) bool {
	return v != data.BillingPrepaid && v != data.BillingPostpaid
}

// fillOffering fills offerings nonce, status, hash and signature.
func (s *Server) fillOffering(offering *data.Offering) error {
	if offering.ID == "" {
		offering.ID = util.NewUUID()
	}
	offering.Nonce = util.NewUUID()
	offering.OfferStatus = data.OfferRegister
	offering.Status = data.MsgUnpublished
	agent := &data.Account{}
	err := s.db.FindByPrimaryKeyTo(agent, offering.Agent)
	if err != nil {
		return err
	}
	offering.Agent = agent.EthAddr
	// TODO: fix this
	offering.BlockNumberUpdated = 1
	hash := data.OfferingHash(offering)
	offering.Hash = data.FromBytes(hash)

	sig, err := agent.Sign(hash)
	if err != nil {
		return err
	}
	offering.Signature = data.FromBytes(sig)
	return nil
}

// handleGetOfferings replies with all offerings or an offering by id.
func (s *Server) handleGetOfferings(w http.ResponseWriter, r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params: []queryParam{{Name: "id", Field: "id"}, {Name: "product", Field: "product"}},
		View:   data.OfferingTable,
	})
}

// Actions that change offerings state.
const (
	PublishOffering    = "publish"
	PopupOffering      = "popup"
	DeactivateOffering = "deactivate"
)

func (s *Server) handlePutOfferingStatus(
	w http.ResponseWriter, r *http.Request, id string) {
	req := &ActionPayload{}
	if !s.parsePayload(w, r, req) {
		return
	}
	s.logger.Info("action ( %v )  request for offering with id: %v recieved.", req.Action, id)
	// TODO once job queue implemented.
}

// handleGetOfferingStatus replies with offerings status by id.
func (s *Server) handleGetOfferingStatus(
	w http.ResponseWriter, r *http.Request, id string) {
	offering := &data.Offering{}
	if !s.findTo(w, offering, id) {
		return
	}
	s.replyStatus(w, offering.Status)
}
