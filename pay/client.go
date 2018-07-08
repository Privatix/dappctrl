package pay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
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

	agentAddr, err := data.ToAddress(ch.Agent)
	if err != nil {
		return nil, err
	}

	offerHash, err := data.ToHash(offer.Hash)
	if err != nil {
		return nil, err
	}

	pscAddrParsed, err := data.ToAddress(pscAddr)
	if err != nil {
		return nil, err
	}

	hash := eth.BalanceProofHash(pscAddrParsed,
		agentAddr, ch.Block, offerHash, big.NewInt(int64(amount)))

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
	pld *paymentPayload, tls bool, timeout uint) error {
	body, err := json.Marshal(pld)
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

	client := http.Client{
		Timeout: time.Duration(timeout) * time.Millisecond,
	}

	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var err srv.Error
		if err := json.NewDecoder(resp.Body).Decode(&err); err != nil {
			return err
		}
		return fmt.Errorf("%s (%d)", err.Message, err.Code)
	}

	return nil
}

// PostCheque sends a payment cheque to a payment server.
func PostCheque(db *reform.DB, channel, pscAddr, pass string,
	amount uint64, tls bool, timeout uint) error {
	pld, err := newPayload(db, channel, pscAddr, pass, amount)
	if err != nil {
		return err
	}
	return postPayload(db, channel, pld, tls, timeout)
}
