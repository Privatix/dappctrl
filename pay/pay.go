package pay

import (
	"net/http"
)

// payload is a balance proof received from a client.
type payload struct {
	AgentAddress    string `json:"agentAddress"`
	OpenBlockNumber uint32 `json:"openBlockNum"`
	OfferingHash    string `json:"offeringHash"`
	Balance         uint64 `json:"balance"`
	BalanceMsgSig   string `json:"balanceMsgSig"`
	ContractAddress string `json:"contractAddress"`
}

// handlePay handles clients balance proof informations.
func (s *Server) handlePay(w http.ResponseWriter, r *http.Request) {
	pld := &payload{}
	if !s.parsePayload(w, r, pld) {
		return
	}
	ch, ok := s.findChannel(w, pld.OfferingHash, pld.AgentAddress,
		pld.OpenBlockNumber)
	if !ok || !s.validateChannelForPayment(w, ch, pld) ||
		!s.updateChannelWithPayment(w, ch, pld) {
		return
	}
	w.WriteHeader(http.StatusOK)
}
