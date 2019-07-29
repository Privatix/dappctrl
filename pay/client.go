package pay

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util/srv"
)

func newPayload(db *reform.DB, channel *data.Channel, pscAddr data.HexString,
	key *ecdsa.PrivateKey, amount uint64) (*paymentPayload, error) {

	var offer data.Offering
	if err := db.FindByPrimaryKeyTo(&offer, channel.Offering); err != nil {
		return nil, err
	}

	pld := &paymentPayload{
		AgentAddress:    channel.Agent,
		OpenBlockNumber: channel.Block,
		OfferingHash:    offer.Hash,
		Balance:         amount,
		ContractAddress: pscAddr,
	}

	agentAddr, err := data.HexToAddress(channel.Agent)
	if err != nil {
		return nil, err
	}

	offerHash, err := data.HexToHash(offer.Hash)
	if err != nil {
		return nil, err
	}

	pscAddrParsed, err := data.HexToAddress(pscAddr)
	if err != nil {
		return nil, err
	}

	hash := eth.BalanceProofHash(pscAddrParsed, agentAddr, channel.Block, offerHash,
		uint64(amount))

	sig, err := crypto.Sign(hash, key)
	if err != nil {
		return nil, err
	}

	pld.BalanceMsgSig = data.FromBytes(sig)

	return pld, nil
}

func postPayload(db *reform.DB, channel *data.Channel, pld *paymentPayload,
	tls bool, timeout uint, pr *proc.Processor,
	sendFunc func(req *http.Request) (*srv.Response, error)) error {
	pldArgs, err := json.Marshal(pld)
	if err != nil {
		return err
	}

	var endp data.Endpoint
	if err := db.FindOneTo(&endp, "channel", channel.ID); err != nil {
		return err
	}

	if endp.PaymentReceiverAddress == nil {
		return fmt.Errorf("no payment addr found for chan %s", channel)
	}
	//TODO: add URL validation and TLS support
	url := *endp.PaymentReceiverAddress

	req, err := srv.NewHTTPRequestWithURL(
		http.MethodPost, url, &srv.Request{Args: pldArgs})

	resp, err := sendFunc(req)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		if resp.Error.Code == errCodeTerminatedService {
			_, err = pr.TerminateChannel(
				channel.ID, data.JobBillingChecker, false)
			if err != nil {
				return err
			}
		}
		return resp.Error
	}

	return nil
}

// PostCheque sends a payment cheque to a payment server.
func PostCheque(db *reform.DB, channel *data.Channel,
	pscAddr data.HexString, key *ecdsa.PrivateKey, amount uint64,
	tls bool, timeout uint, pr *proc.Processor) error {
	pld, err := newPayload(db, channel, pscAddr, key, amount)
	if err != nil {
		return err
	}
	return postPayload(db, channel, pld, tls, timeout, pr, srv.Send)
}
