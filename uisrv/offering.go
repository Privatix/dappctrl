package uisrv

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/proc/worker"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

// Actions for Agent that change offerings state.
const (
	PublishOffering    = "publish"
	PopupOffering      = "popup"
	DeactivateOffering = "deactivate"
)

// Actions for Client that change offerings state.
const (
	AcceptOffering = "accept"
)

const (
	clientGetOfferFilter = `offer_status in ('registered', 'popped_up')
                                AND status = 'msg_channel_published'
				AND NOT is_local
				AND offerings.current_supply > 0
                                AND offerings.agent NOT IN
                                (SELECT eth_addr
				   FROM accounts)
				ORDER BY block_number_updated DESC`
	clientGetOfferFilterByID = "id = $1 AND " + clientGetOfferFilter
	agentGetOfferFilter      = `offerings.agent IN
                               (SELECT eth_addr
                                  FROM accounts)
                               AND (SELECT in_use
                                      FROM accounts
                                     WHERE eth_addr = offerings.agent)
			       ORDER BY block_number_updated DESC`
)

type clientPreChannelCreateData struct {
	Account  string `json:"account"`
	Offering string `json:"offering"`
	GasPrice uint64 `json:"gasPrice"`
}

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

// handleOfferings calls appropriate handler by scanning incoming request.
func (s *Server) handleClientOfferings(w http.ResponseWriter, r *http.Request) {
	if id := idFromStatusPath(clientOfferingsPath, r.URL.Path); id != "" {
		if r.Method == http.MethodPut {
			s.handlePutClientOfferingStatus(w, r, id)
			return
		}
	} else {
		if r.Method == http.MethodGet {
			s.handleGetClientOfferings(w, r)
			return
		}
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handlePostOffering creates offering.
func (s *Server) handlePostOffering(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.Add("method", "handlePostOffering")

	offering := &data.Offering{}
	if !s.parseOfferingPayload(logger, w, r, offering) {
		return
	}
	err := s.fillOffering(offering)
	if err != nil {
		s.replyInvalidRequest(logger, w)
		return
	}
	if !s.insert(logger, w, offering) {
		return
	}
	s.replyEntityCreated(logger, w, offering.ID)
}

// handlePutOffering updates offering.
func (s *Server) handlePutOffering(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.Add("method", "handlePutOffering")

	offering := &data.Offering{}
	if !s.parseOfferingPayload(logger, w, r, offering) {
		return
	}
	err := s.fillOffering(offering)
	if err != nil {
		s.replyUnexpectedErr(logger, w)
		return
	}
	if err := s.db.Update(offering); err != nil {
		s.logger.Warn(fmt.Sprintf("failed to update offering: %v", err))
		s.replyUnexpectedErr(logger, w)
		return
	}
	s.replyEntityUpdated(logger, w, offering.ID)
}

func (s *Server) parseOfferingPayload(logger log.Logger,
	w http.ResponseWriter, r *http.Request, offering *data.Offering) bool {
	if !s.parsePayload(logger, w, r, offering) ||
		validate.Struct(offering) != nil ||
		invalidUnitType(offering.UnitType) ||
		invalidBillingType(offering.BillingType) {
		s.replyInvalidRequest(logger, w)
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
		offering.OfferStatus = data.OfferEmpty
	} else {
		offering.OfferStatus = data.OfferRegistered
	}

	agent := &data.Account{}
	if err := s.db.FindByPrimaryKeyTo(agent, offering.Agent); err != nil {
		return err
	}

	offering.Status = data.MsgUnpublished
	offering.Agent = agent.EthAddr
	offering.BlockNumberUpdated = 1
	offering.CurrentSupply = offering.Supply
	// TODO: remove once prepaid is implemented.
	offering.BillingType = data.BillingPostpaid

	return nil
}

// handleGetClientOfferings replies with all active offerings
// available to the client.
func (s *Server) handleGetClientOfferings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var filtering string
	if r.FormValue("id") == "" {
		filtering = clientGetOfferFilter
	}

	s.handleGetResources(w, r, &getConf{
		Params: []queryParam{
			{Name: "minUnitPrice", Field: "unit_price", Op: ">="},
			{Name: "maxUnitPrice", Field: "unit_price", Op: "<="},
			{Name: "country", Field: "country", Op: "in"},
			{Name: "agent", Field: "agent"},
			{Name: "id", Field: "id"},
		},
		View: data.OfferingTable,
		FilteringSQL: filteringSQL{
			SQL:      filtering,
			JoinWith: "AND",
		},
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
		View: data.OfferingTable,
		FilteringSQL: filteringSQL{
			SQL:      agentGetOfferFilter,
			JoinWith: "AND",
		},
	})
}

// OfferingPutPayload offering status update payload.
type OfferingPutPayload struct {
	Action   string `json:"action"`
	Account  string `json:"account"`
	GasPrice uint64 `json:"gasPrice"`
}

func (s *Server) handlePutOfferingStatus(
	w http.ResponseWriter, r *http.Request, id string) {
	logger := s.logger.Add("method", "handlePutOfferingStatus", "id", id)

	req := &OfferingPutPayload{}
	if !s.parsePayload(logger, w, r, req) {
		return
	}

	var jobType string
	if req.Action == PublishOffering {
		jobType = data.JobAgentPreOfferingMsgBCPublish
	} else if req.Action == PopupOffering {
		jobType = data.JobAgentPreOfferingPopUp
	} else if req.Action == DeactivateOffering {
		jobType = data.JobAgentPreOfferingDelete
	} else {
		s.replyInvalidAction(logger, w)
		return
	}

	if !s.findTo(logger, w, &data.Offering{}, id) {
		return
	}

	logger.Info(fmt.Sprintf(
		"action ( %v )  request for offering with id: %v recieved.",
		req.Action, id))

	gasPrice := req.GasPrice

	if req.GasPrice == 0 {
		val, ok := s.defaultGasPrice(logger, w)
		if !ok {
			return
		}
		gasPrice = val
	}

	dataJSON, err := json.Marshal(&data.JobPublishData{GasPrice: gasPrice})
	if err != nil {
		s.logger.Error(
			fmt.Sprintf("failed to marshal job data: %v", err))
		s.replyUnexpectedErr(logger, w)
		return
	}

	if err := s.queue.Add(&data.Job{
		Type:        jobType,
		RelatedType: data.JobOffering,
		RelatedID:   id,
		CreatedBy:   data.JobUser,
		Data:        dataJSON,
	}); err != nil {
		s.logger.Error(fmt.Sprintf("failed to add %s: %v",
			data.JobAgentPreOfferingMsgBCPublish, err))
		s.replyUnexpectedErr(logger, w)
	}

	w.WriteHeader(http.StatusOK)
}

// ClientOfferingPutPayload offering status update payload for clients.
type ClientOfferingPutPayload struct {
	Action   string `json:"action"`
	Account  string `json:"account"`
	GasPrice uint64 `json:"gasPrice"`
	Deposit  uint64 `json:"deposit"`
}

func (s *Server) handlePutClientOfferingStatus(
	w http.ResponseWriter, r *http.Request, id string) {
	logger := s.logger.Add("method", "handlePutClientOfferingStatus",
		"id", id)

	req := new(ClientOfferingPutPayload)
	if !s.parsePayload(logger, w, r, req) {
		return
	}

	logger = logger.Add("payload", *req)

	if req.Action != AcceptOffering {
		s.replyInvalidAction(logger, w)
		return
	}

	offer := new(data.Offering)
	acc := new(data.Account)

	if !s.selectOneTo(logger, w, offer, "WHERE "+clientGetOfferFilterByID, id) {
		return
	}

	minDeposit := data.MinDeposit(offer)

	if req.Deposit == 0 {
		req.Deposit = minDeposit
	} else if req.Deposit < minDeposit {
		s.replyInvalidAction(logger, w)
		return
	}

	if !s.findTo(logger, w, acc, req.Account) {
		return
	}

	logger.Info(
		fmt.Sprintf("action ( %v )  request for offering with id:"+
			" %v received.", req.Action, id))

	dataJSON, err := json.Marshal(&worker.ClientPreChannelCreateData{
		GasPrice: req.GasPrice, Offering: offer.ID, Account: acc.ID,
		Deposit: req.Deposit})
	if err != nil {
		logger.Error(fmt.Sprintf("failed to marshal job data: %v", err))
		s.replyUnexpectedErr(logger, w)
		return
	}
	if err := s.queue.Add(&data.Job{
		Type:        data.JobClientPreChannelCreate,
		RelatedType: data.JobChannel,
		RelatedID:   util.NewUUID(),
		CreatedBy:   data.JobUser,
		Data:        dataJSON,
	}); err != nil {
		s.logger.Error(fmt.Sprintf("failed to add %s: %v",
			data.JobClientPreChannelCreate, err))
		s.replyUnexpectedErr(logger, w)
	}

	w.WriteHeader(http.StatusOK)
}

// handleGetOfferingStatus replies with offerings status by id.
func (s *Server) handleGetOfferingStatus(
	w http.ResponseWriter, r *http.Request, id string) {
	logger := s.logger.Add("method", "handleGetOfferingStatus")

	offering := &data.Offering{}
	if !s.findTo(logger, w, offering, id) {
		return
	}
	s.replyStatus(logger, w, offering.Status)
}
