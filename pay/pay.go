package pay

import (
	"fmt"
	"net/http"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/srv"
)

// paymentPayload is a balance proof received from a client.
type paymentPayload struct {
	AgentAddress    data.HexString    `json:"agentAddress"`
	OpenBlockNumber uint32            `json:"openBlockNum"`
	OfferingHash    data.HexString    `json:"offeringHash"`
	Balance         uint64            `json:"balance"`
	BalanceMsgSig   data.Base64String `json:"balanceMsgSig"`
	ContractAddress data.HexString    `json:"contractAddress"`
}

// handlePay handles clients balance proof informations.
func (s *Server) handlePay(
	w http.ResponseWriter, r *http.Request, ctx *srv.Context) {
	logger := s.logger.Add("method", "handlePay", "sender", r.RemoteAddr)

	payload := &paymentPayload{}
	if !s.ParseRequest(logger, w, r, payload) {
		return
	}
	logger = logger.Add("payload", *payload)

	ch, ok := s.findChannel(logger, w,
		payload.OfferingHash,
		payload.AgentAddress, payload.OpenBlockNumber)

	logger = logger.Add("channel", ch)

	if ok && s.isServiceTerminated(logger, w, ch) {
		return
	}

	if !ok || !s.validateChannelForPayment(logger, w, ch, payload) ||
		!s.updateChannelWithPayment(logger, w, ch, payload) {
		return
	}

	s.RespondResult(logger, w, struct{}{})

	logger.Info(fmt.Sprintf("received payment: %d, from: %s", payload.Balance, ch.Client))
}
