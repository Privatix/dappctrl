package pay

import (
	"encoding/json"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
)

// serverError is a payment server error.
type serverError struct {
	// Code is a status code.
	Code int `json:"code"`
	// Message is a description of the error.
	Message string `json:"message"`
}

var (
	errInvalidPayload = &serverError{
		Message: "",
	}
	errNoChannel = &serverError{
		Message: "Channel is not found",
	}
	errUnexpected = &serverError{
		Message: "An unexpected error occurred",
	}
	errChannelClosed = &serverError{
		Message: "Channel is closed",
	}
	errInvalidAmount = &serverError{
		Message: "Invalid balance amount",
	}
	errInvalidSignature = &serverError{
		Message: "Client signature does not match",
	}
)

func (s *Server) findChannel(w http.ResponseWriter,
	offeringHash string,
	agentAddr string,
	block uint32) (*data.Channel, bool) {

	ch := &data.Channel{}

	tail := "INNER JOIN offerings ON offerings.hash=$1 WHERE channels.agent=$2 AND channels.block=$3"
	err := s.db.SelectOneTo(ch, tail, offeringHash, agentAddr, block)
	if err != nil {
		s.replyErr(w, http.StatusUnauthorized, errNoChannel)
		return nil, false
	}

	return ch, true
}

func (s *Server) validateChannelState(w http.ResponseWriter,
	ch *data.Channel) bool {
	if ch.ChannelStatus != data.ChannelActive {
		s.replyErr(w, http.StatusUnauthorized, errChannelClosed)
		return false
	}
	return true
}

func (s *Server) validateAmount(w http.ResponseWriter,
	ch *data.Channel, pld *payload) bool {
	if pld.Balance <= ch.ReceiptBalance || pld.Balance > ch.TotalDeposit {
		s.replyErr(w, http.StatusBadRequest, errInvalidAmount)
		return false
	}
	return true
}

func (s *Server) verifySignature(w http.ResponseWriter,
	ch *data.Channel, pld *payload) bool {

	client := &data.User{}
	if s.db.FindOneTo(client, "eth_addr", ch.Client) != nil {
		s.replyErr(w, http.StatusInternalServerError, errUnexpected)
		return false
	}

	pub, err := data.ToBytes(client.PublicKey)
	if err != nil {
		s.replyErr(w, http.StatusInternalServerError, errUnexpected)
		return false
	}

	sig, err := data.ToBytes(pld.BalanceMsgSig)
	if err != nil {
		s.replyErr(w, http.StatusInternalServerError, errUnexpected)
		return false
	}

	pscAddr, err := data.ToAddress(pld.ContractAddress)
	if err != nil {
		s.replyErr(w, http.StatusInternalServerError, errUnexpected)
		return false
	}

	agentAddr, err := data.ToAddress(ch.Agent)
	if err != nil {
		s.replyErr(w, http.StatusInternalServerError, errUnexpected)
		return false
	}

	offeringHash, err := data.ToHash(pld.OfferingHash)
	if err != nil {
		s.replyErr(w, http.StatusInternalServerError, errUnexpected)
		return false
	}

	hash := eth.BalanceProofHash(pscAddr, agentAddr,
		pld.OpenBlockNumber, offeringHash, big.NewInt(int64(pld.Balance)))

	if !crypto.VerifySignature(pub, hash, sig[:len(sig)-1]) {
		s.replyErr(w, http.StatusBadRequest, errInvalidSignature)
		return false
	}
	return true
}

func (s *Server) validateChannelForPayment(w http.ResponseWriter,
	ch *data.Channel, pld *payload) bool {
	return s.validateChannelState(w, ch) &&
		s.validateAmount(w, ch, pld) &&
		s.verifySignature(w, ch, pld)
}

func (s *Server) updateChannelWithPayment(w http.ResponseWriter,
	ch *data.Channel, pld *payload) bool {
	ch.ReceiptBalance = pld.Balance
	ch.ReceiptSignature = &pld.BalanceMsgSig
	if err := s.db.Update(ch); err != nil {
		s.logger.Warn("failed to update channel: %v", err)
		s.replyErr(w, http.StatusInternalServerError, errUnexpected)
		return false
	}
	return true
}

func (s *Server) parsePayload(w http.ResponseWriter,
	r *http.Request, v interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		s.logger.Warn("failed to parse request body: %v", err)
		s.replyErr(w, http.StatusBadRequest, errInvalidPayload)
		return false
	}
	return true
}

// replyErr writes error to reponse.
func (s *Server) replyErr(w http.ResponseWriter, status int, reply *serverError) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(reply); err != nil {
		s.logger.Warn("failed to marshal error reply to json: %v", err)
	}
}
