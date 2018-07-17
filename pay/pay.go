package pay

import (
	"net/http"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/srv"
)

// paymentPayload is a balance proof received from a client.
type paymentPayload struct {
	AgentAddress    string `json:"agentAddress"`
	OpenBlockNumber uint32 `json:"openBlockNum"`
	OfferingHash    string `json:"offeringHash"`
	Balance         uint64 `json:"balance"`
	BalanceMsgSig   string `json:"balanceMsgSig"`
	ContractAddress string `json:"contractAddress"`
}

// handlePay handles clients balance proof informations.
func (s *Server) handlePay(
	w http.ResponseWriter, r *http.Request, ctx *srv.Context) {
	payload := &paymentPayload{}
	if _, ok := s.ParseRequest(w, r, payload); !ok {
		return
	}
	ch, ok := s.findChannel(w,
		payload.OfferingHash,
		payload.AgentAddress, payload.OpenBlockNumber)

	if !ok || !s.validateChannelForPayment(w, ch, payload) ||
		!s.updateChannelWithPayment(w, ch, payload) {
		return
	}

	s.RespondResult(w, struct{}{})

	clientHex, err := data.FromBase64ToHex(ch.Client)
	if err != nil {
		s.Logger().Error("failed to decode client addr: %v", err)
		return
	}

	s.Logger().Info(
		"received payment: %d, from: %s", payload.Balance, clientHex)
}
