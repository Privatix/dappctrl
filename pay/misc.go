package pay

import (
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/util/srv"
)

// Codes for unauthorized replies.
const (
	errCodeNoChannel     = 1
	errCodeClosedChannel = iota
)

// Codes for bad request replies.
const (
	errCodeNonParsablePayload = 1
	errCodeInvalidBalance     = iota
	errCodeInvalidSignature   = iota
)

var errUnexpected = &srv.Error{
	Status:  http.StatusBadRequest,
	Code:    errCodeInvalidSignature,
	Message: "Client signature does not match",
}

func (s *Server) findChannel(w http.ResponseWriter, offeringHash string,
	agentAddr string, block uint32) (*data.Channel, bool) {

	channel := &data.Channel{}
	tail := `INNER JOIN offerings
		  ON offerings.hash=$1
		  WHERE channels.agent=$2 AND channels.block=$3`

	err := s.db.SelectOneTo(channel, tail, offeringHash, agentAddr, block)
	if err != nil {
		s.RespondError(w, &srv.Error{
			Status:  http.StatusUnauthorized,
			Message: "Channel is not found",
			Code:    errCodeNoChannel,
		})
		return nil, false
	}

	return channel, true
}

func (s *Server) validateChannelState(w http.ResponseWriter,
	ch *data.Channel) bool {
	if ch.ChannelStatus != data.ChannelActive {
		s.RespondError(w, &srv.Error{
			Status:  http.StatusUnauthorized,
			Message: "Channel is closed",
			Code:    errCodeClosedChannel,
		})
		return false
	}
	return true
}

func (s *Server) verifySignature(w http.ResponseWriter,
	ch *data.Channel, pld *paymentPayload) bool {

	client := &data.User{}
	if err := s.db.FindOneTo(client, "eth_addr", ch.Client); err != nil {
		s.Logger().Warn("could not find client with addr %v: %v",
			ch.Client, err)
		s.RespondError(w, errUnexpected)
		return false
	}

	pub, err := data.ToBytes(client.PublicKey)
	if err != nil {
		s.Logger().Error("could not decode public key: %v", err)
		s.RespondError(w, errUnexpected)
		return false
	}

	sig, err := data.ToBytes(pld.BalanceMsgSig)
	if err != nil {
		s.Logger().Error("could not decode signature: %v", err)
		s.RespondError(w, errUnexpected)
		return false
	}

	pscAddr, err := data.HexToAddress(pld.ContractAddress)
	if err != nil {
		s.Logger().Error("could not parse contract addr: %v", err)
		s.RespondError(w, errUnexpected)
		return false
	}

	agentAddr, err := data.HexToAddress(ch.Agent)
	if err != nil {
		s.Logger().Error("could not parse agent addr: %v", err)
		s.RespondError(w, errUnexpected)
		return false
	}

	offeringHash, err := data.ToHash(pld.OfferingHash)
	if err != nil {
		s.Logger().Error("could not parse offering hash: %v", err)
		s.RespondError(w, errUnexpected)
		return false
	}

	hash := eth.BalanceProofHash(pscAddr, agentAddr,
		pld.OpenBlockNumber, offeringHash, big.NewInt(int64(pld.Balance)))

	if !crypto.VerifySignature(pub, hash, sig[:len(sig)-1]) {
		s.RespondError(w, &srv.Error{
			Status:  http.StatusBadRequest,
			Code:    errCodeInvalidSignature,
			Message: "Client signature does not match",
		})
		return false
	}
	return true
}

func (s *Server) validateChannelForPayment(w http.ResponseWriter,
	ch *data.Channel, pld *paymentPayload) bool {
	return s.validateChannelState(w, ch) &&
		s.verifySignature(w, ch, pld)
}

func (s *Server) updateChannelWithPayment(w http.ResponseWriter,
	ch *data.Channel, pld *paymentPayload) bool {
	ch.ReceiptBalance = pld.Balance
	ch.ReceiptSignature = &pld.BalanceMsgSig
	ret, err := s.db.Exec(`
		UPDATE channels set receipt_balance=$1, receipt_signature=$2
		 WHERE receipt_balance<$1 AND total_deposit>=$1 AND id=$3`,
		pld.Balance, pld.BalanceMsgSig, ch.ID)
	if err != nil {
		s.Logger().Warn("failed to update channel: %v", err)
		s.RespondError(w, errUnexpected)
		return false
	}
	affected, err := ret.RowsAffected()
	if err != nil {
		s.Logger().Error("could not get rows affected number: %v", err)
		s.RespondError(w, errUnexpected)
		return false
	}
	if affected == 0 {
		s.RespondError(w, &srv.Error{
			Status:  http.StatusBadRequest,
			Code:    errCodeInvalidBalance,
			Message: "Invalid balance amount",
		})
		return false
	}
	return true
}
