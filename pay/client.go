package pay

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util/srv"
)

func newPayload(db *reform.DB, channel,
	pscAddr, pass string, amount uint64) (*paymentPayload, error) {
	var ch data.Channel
	if err := db.FindByPrimaryKeyTo(&ch, channel); err != nil {
		return nil, err
	}

	var offer data.Offering
	if err := db.FindByPrimaryKeyTo(&offer, ch.Offering); err != nil {
		return nil, err
	}

	var client data.Account
	if err := db.FindOneTo(&client, "eth_addr", ch.Client); err != nil {
		return nil, err
	}

	pld := &paymentPayload{
		AgentAddress:    ch.Agent,
		OpenBlockNumber: ch.Block,
		OfferingHash:    offer.Hash,
		Balance:         amount,
		ContractAddress: pscAddr,
	}

	agentAddr, err := data.HexToAddress(ch.Agent)
	if err != nil {
		return nil, err
	}

	offerHash, err := data.ToHash(offer.Hash)
	if err != nil {
		return nil, err
	}

	pscAddrParsed, err := data.HexToAddress(pscAddr)
	if err != nil {
		return nil, err
	}

	hash := eth.BalanceProofHash(pscAddrParsed,
		agentAddr, ch.Block, offerHash, new(big.Int).SetUint64(amount))

	key, err := data.ToPrivateKey(client.PrivateKey, pass)
	if err != nil {
		return nil, err
	}

	sig, err := crypto.Sign(hash, key)
	if err != nil {
		return nil, err
	}

	pld.BalanceMsgSig = data.FromBytes(sig)

	return pld, nil
}

func postPayload(db *reform.DB, channel string,
	pld *paymentPayload, tls bool, timeout uint, pr *proc.Processor) error {
	pldArgs, err := json.Marshal(pld)
	if err != nil {
		return err
	}

	var endp data.Endpoint
	if err := db.FindOneTo(&endp, "channel", channel); err != nil {
		return err
	}

	if endp.PaymentReceiverAddress == nil {
		return fmt.Errorf("no payment addr found for chan %s", channel)
	}
	//TODO: add URL validation and TLS support
	url := *endp.PaymentReceiverAddress

	req, err := srv.NewHTTPRequestWithURL(
		http.MethodPost, url, &srv.Request{Args: pldArgs})

	resp, err := srv.Send(req)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		if resp.Error.Code == errCodeTerminatedService {
			_, err = pr.TerminateChannel(
				channel, data.JobBillingChecker, false)
			if err != nil {
				return err
			}
		}
		return fmt.Errorf("%s (%d)", resp.Error.Message, resp.Error.Code)
	}

	return nil
}

// PostCheque sends a payment cheque to a payment server.
func PostCheque(db *reform.DB, channel, pscAddr, pass string,
	amount uint64, tls bool, timeout uint, pr *proc.Processor) error {
	pld, err := newPayload(db, channel, pscAddr, pass, amount)
	if err != nil {
		return err
	}
	return postPayload(db, channel, pld, tls, timeout, pr)
}
