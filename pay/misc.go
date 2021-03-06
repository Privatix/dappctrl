package pay

import (
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/util/srv"
)

// Codes for unauthorized replies.
const (
	errCodeNoChannel = iota + 1
	errCodeClosedChannel
	errCodeTerminatedService
)

// Codes for bad request replies.
const (
	errCodeNonParsablePayload = iota + 1
	errCodeInvalidBalance
	errCodeInvalidSignature
	ErrCodeEqualBalance
)

var errUnexpected = &srv.Error{
	Status:  http.StatusBadRequest,
	Code:    errCodeInvalidSignature,
	Message: "Unexpected error occurred",
}

func (s *Server) findChannel(logger log.Logger,
	w http.ResponseWriter, offeringHash, agentAddr data.HexString,
	block uint32) (*data.Channel, bool) {

	channel := &data.Channel{}
	tail := `INNER JOIN offerings
		  ON offerings.hash=$1
		  WHERE channels.agent=$2 AND channels.block=$3`

	err := s.db.SelectOneTo(channel, tail, offeringHash, agentAddr, block)
	if err != nil {
		logger.Warn("channel is not found")
		s.RespondError(logger, w, &srv.Error{
			Status:  http.StatusUnauthorized,
			Message: "Channel is not found",
			Code:    errCodeNoChannel,
		})
		return nil, false
	}

	return channel, true
}

func (s *Server) validateChannelState(logger log.Logger,
	w http.ResponseWriter, ch *data.Channel) bool {
	if ch.ChannelStatus != data.ChannelActive {
		s.RespondError(logger, w, &srv.Error{
			Status:  http.StatusUnauthorized,
			Message: "Channel is closed",
			Code:    errCodeClosedChannel,
		})
		logger.Warn("channel is closed")
		return false
	}
	return true
}

func (s *Server) isServiceTerminated(logger log.Logger,
	w http.ResponseWriter, ch *data.Channel) bool {
	if ch.ServiceStatus == data.ServiceTerminated {
		s.RespondError(logger, w, &srv.Error{
			Status:  http.StatusUnauthorized,
			Message: "Service is terminated",
			Code:    errCodeTerminatedService,
		})
		logger.Warn("service is terminated")
		return true
	}
	return false
}

func (s *Server) verifySignature(logger log.Logger,
	w http.ResponseWriter, ch *data.Channel, pld *paymentPayload) bool {

	client := &data.User{}
	if err := s.db.FindOneTo(client, "eth_addr", ch.Client); err != nil {
		logger.Warn("could not find client: " + err.Error())
		s.RespondError(logger, w, errUnexpected)
		return false
	}

	pub, err := data.ToBytes(client.PublicKey)
	if err != nil {
		logger.Error("could not decode public key: " + err.Error())
		s.RespondError(logger, w, errUnexpected)
		return false
	}

	sig, err := data.ToBytes(pld.BalanceMsgSig)
	if err != nil {
		logger.Error("could not decode signature: " + err.Error())
		s.RespondError(logger, w, errUnexpected)
		return false
	}

	pscAddr, err := data.HexToAddress(pld.ContractAddress)
	if err != nil {
		logger.Error("could not parse contract addr: " + err.Error())
		s.RespondError(logger, w, errUnexpected)
		return false
	}

	agentAddr, err := data.HexToAddress(ch.Agent)
	if err != nil {
		logger.Error("could not parse agent addr: " + err.Error())
		s.RespondError(logger, w, errUnexpected)
		return false
	}

	offeringHash, err := data.HexToHash(pld.OfferingHash)
	if err != nil {
		logger.Error("could not parse offering hash: " + err.Error())
		s.RespondError(logger, w, errUnexpected)
		return false
	}

	hash := eth.BalanceProofHash(pscAddr, agentAddr, pld.OpenBlockNumber,
		offeringHash, uint64(pld.Balance))

	if !crypto.VerifySignature(pub, hash, sig[:len(sig)-1]) {
		s.RespondError(logger, w, &srv.Error{
			Status:  http.StatusBadRequest,
			Code:    errCodeInvalidSignature,
			Message: "Client signature does not match",
		})
		logger.Warn("client signature does not match")
		return false
	}
	return true
}

func (s *Server) validateChannelForPayment(logger log.Logger,
	w http.ResponseWriter, ch *data.Channel, pld *paymentPayload) bool {
	return s.validateChannelState(logger, w, ch) &&
		s.verifySignature(logger, w, ch, pld)
}

func (s *Server) updateChannelWithPayment(logger log.Logger,
	w http.ResponseWriter, ch *data.Channel, pld *paymentPayload) bool {
	// Check receipt balance.
	if ch.ReceiptBalance == pld.Balance {
		s.RespondError(logger, w, &srv.Error{
			Status:  http.StatusBadRequest,
			Code:    ErrCodeEqualBalance,
			Message: "Balance amount is equal to current balance",
		})
		return false
	}

	ch.ReceiptBalance = pld.Balance
	ch.ReceiptSignature = &pld.BalanceMsgSig
	ret, err := s.db.Exec(`
		UPDATE channels set receipt_balance=$1, receipt_signature=$2
		 WHERE receipt_balance<$1 AND total_deposit>=$1 AND id=$3`,
		pld.Balance, pld.BalanceMsgSig, ch.ID)
	if err != nil {
		logger.Warn("failed to update channel: " + err.Error())
		s.RespondError(logger, w, errUnexpected)
		return false
	}
	affected, err := ret.RowsAffected()
	if err != nil {
		logger.Error("could not get rows affected number: " + err.Error())
		s.RespondError(logger, w, errUnexpected)
		return false
	}
	if affected == 0 {
		s.RespondError(logger, w, &srv.Error{
			Status:  http.StatusBadRequest,
			Code:    errCodeInvalidBalance,
			Message: "Invalid balance amount",
		})
		return false
	}
	return true
}
