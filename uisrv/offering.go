package uisrv

import (
	"encoding/json"
	"net/http"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

// handleOfferings calls appropriate handler by scanning incoming request.
func (s *Server) handleOfferings(w http.ResponseWriter, r *http.Request) {
	if id := idFromStatusPath(offeringsPath, r.URL.Path); id != "" {
		if r.Method == http.MethodPut {
			s.handlePutOfferingStatus(w, r, id)
			return
		}
		if r.Method == http.MethodGet {
			s.handleGetOfferingStatus(w, r, id)
			return
		}
	} else {
		if r.Method == http.MethodPost {
			s.handlePostOffering(w, r)
			return
		}
		if r.Method == http.MethodPut {
			s.handlePutOffering(w, r)
			return
		}
		if r.Method == http.MethodGet {
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

	agent := &data.Account{}
	if err := s.db.FindByPrimaryKeyTo(agent, offering.Agent); err != nil {
		return err
	}

	offering.OfferStatus = data.OfferRegister
	offering.Status = data.MsgUnpublished
	offering.Agent = agent.EthAddr
	offering.BlockNumberUpdated = 1

	return nil
}

// handleGetClientOfferings replies with all active offerings
// available to the client.
func (s *Server) handleGetClientOfferings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	s.handleGetResources(w, r, &getConf{
		Params: []queryParam{
			{Name: "minUnitPrice", Field: "unit_price", Op: ">="},
			{Name: "maxUnitPrice", Field: "unit_price", Op: "<="},
			{Name: "country", Field: "country", Op: "in"},
		},
		View:         data.OfferingTable,
		FilteringSQL: `offer_status = 'register' AND status = 'msg_channel_published' AND NOT is_local AND offerings.agent NOT IN (SELECT eth_addr FROM accounts)`,
	})
}

// handleGetOfferings replies with all offerings or an offering by id
// available to the agent.
func (s *Server) handleGetOfferings(w http.ResponseWriter, r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params: []queryParam{
			{Name: "id", Field: "id"},
			{Name: "product", Field: "product"},
			{Name: "offerStatus", Field: "offer_status"},
		},
		View:         data.OfferingTable,
		FilteringSQL: `offerings.agent IN (SELECT eth_addr FROM accounts) AND (SELECT in_use FROM accounts WHERE eth_addr = offerings.agent)`,
	})
}

// Actions that change offerings state.
const (
	PublishOffering    = "publish"
	PopupOffering      = "popup"
	DeactivateOffering = "deactivate"
)

// OfferingPutPayload offering status update payload.
type OfferingPutPayload struct {
	Action   string `json:"action"`
	GasPrice uint64 `json:"gasPrice"`
}

func (s *Server) handlePutOfferingStatus(
	w http.ResponseWriter, r *http.Request, id string) {
	req := &OfferingPutPayload{}
	if !s.parsePayload(w, r, req) {
		return
	}
	// TODO: popup, deactivate
	if req.Action != PublishOffering {
		s.replyInvalidAction(w)
		return
	}
	if !s.findTo(w, &data.Offering{}, id) {
		return
	}
	s.logger.Info("action ( %v )  request for offering with id: %v recieved.", req.Action, id)

	dataJSON, err := json.Marshal(&data.JobPublishData{GasPrice: req.GasPrice})
	if err != nil {
		s.logger.Error("failed to marshal job data: %v", err)
		s.replyUnexpectedErr(w)
		return
	}

	if err := s.queue.Add(&data.Job{
		Type:        data.JobAgentPreOfferingMsgBCPublish,
		RelatedType: data.JobOfferring,
		RelatedID:   id,
		CreatedBy:   data.JobUser,
		Data:        dataJSON,
	}); err != nil {
		s.logger.Error("failed to add %s: %v",
			data.JobAgentPreOfferingMsgBCPublish, err)
		s.replyUnexpectedErr(w)
	}

	w.WriteHeader(http.StatusOK)
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
