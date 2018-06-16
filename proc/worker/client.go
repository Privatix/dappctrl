package worker

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/AlekSi/pointer"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/messages"
	"github.com/privatix/dappctrl/messages/ept"
	"github.com/privatix/dappctrl/statik"
	"github.com/privatix/dappctrl/util"
)

func (w *Worker) clientPreChannelCreateCheckDeposit(
	acc *data.Account, offer *data.Offering) (uint64, error) {
	addr, err := data.ToAddress(acc.EthAddr)
	if err != nil {
		return 0, err
	}

	amount, err := w.ethBack.PSCBalanceOf(&bind.CallOpts{}, addr)
	if err != nil {
		return 0, err
	}

	deposit := offer.UnitPrice*offer.MinUnits + offer.SetupPrice
	if amount.Uint64() < deposit {
		return 0, ErrNotEnoughBalance
	}

	return deposit, nil
}

func (w *Worker) clientPreChannelCreateCheckSupply(
	offer *data.Offering, offerHash common.Hash) error {
	supply, err := w.ethBack.PSCOfferingSupply(&bind.CallOpts{}, offerHash)
	if err != nil {
		return err
	}

	if supply == 0 {
		return ErrNoSupply
	}

	return nil
}

func (w *Worker) clientPreChannelCreateSaveTX(
	job *data.Job, acc *data.Account, offer *data.Offering,
	offerHash common.Hash, deposit uint64) error {
	agentAddr, err := data.ToAddress(offer.Agent)
	if err != nil {
		return err
	}

	key, err := w.key(acc.PrivateKey)
	if err != nil {
		return err
	}

	// TODO: set gas limit from conf
	tx, err := w.ethBack.PSCCreateChannel(bind.NewKeyedTransactor(key),
		agentAddr, offerHash, big.NewInt(int64(deposit)))
	if err != nil {
		return err
	}

	if err := w.saveEthTX(job, tx, "CreateChannel", data.JobChannel,
		job.RelatedID, acc.EthAddr, offer.Agent); err != nil {
		return err
	}

	ch := data.Channel{
		ID:            job.RelatedID,
		Agent:         offer.Agent,
		Client:        acc.EthAddr,
		Offering:      offer.ID,
		Block:         0,
		ChannelStatus: data.ChannelPending,
		ServiceStatus: data.ServicePending,
		TotalDeposit:  deposit,
	}
	return data.Insert(w.db.Querier, &ch)
}

// ClientPreChannelCreateData is a job data for ClientPreChannelCreate.
type ClientPreChannelCreateData struct {
	Account  string `json:"account"`
	Oferring string `json:"offering"`
}

// ClientPreChannelCreate triggers a channel creation.
func (w *Worker) ClientPreChannelCreate(job *data.Job) error {
	var jdata ClientPreChannelCreateData
	if err := parseJobData(job, &jdata); err != nil {
		return err
	}

	var acc data.Account
	err := data.FindByPrimaryKeyTo(w.db.Querier, &acc, jdata.Account)
	if err != nil {
		return err
	}

	var offer data.Offering
	err = data.FindByPrimaryKeyTo(w.db.Querier, &offer, jdata.Oferring)
	if err != nil {
		return err
	}

	deposit, err := w.clientPreChannelCreateCheckDeposit(&acc, &offer)
	if err != nil {
		return err
	}

	offerHash, err := data.ToHash(offer.Hash)
	if err != nil {
		return err
	}

	err = w.clientPreChannelCreateCheckSupply(&offer, offerHash)
	if err != nil {
		return err
	}

	return w.clientPreChannelCreateSaveTX(
		job, &acc, &offer, offerHash, deposit)
}

// ClientAfterChannelCreate activates channel and triggers endpoint retrieval.
func (w *Worker) ClientAfterChannelCreate(job *data.Job) error {
	var ch data.Channel
	err := data.FindByPrimaryKeyTo(w.db.Querier, &ch, job.RelatedID)
	if err != nil {
		return err
	}

	ch.ChannelStatus = data.ChannelActive
	if err = data.Save(w.db.Querier, &ch); err != nil {
		return err
	}

	go func() {
		ep, err := w.somc.WaitForEndpoint(job.RelatedID)
		if err != nil {
			w.logger.Error("failed to get endpoint for chan %s: %s",
				ch.ID, err)
			return
		}

		err = w.addJobWithData(data.JobClientPreEndpointMsgSOMCGet,
			data.JobChannel, ch.ID, ep)
		if err != nil {
			w.logger.Error("failed to add "+
				"JobClientPreEndpointMsgSOMCGet job: %s", err)
		}
	}()

	return nil
}

func (w *Worker) decodeEndpoint(
	ch *data.Channel, sealed []byte) (*ept.Message, error) {
	var client data.Account
	err := data.FindOneTo(w.db.Querier, &client, "eth_addr", ch.Client)
	if err != nil {
		return nil, err
	}

	var agent data.User
	err = data.FindOneTo(w.db.Querier, &agent, "eth_addr", ch.Agent)
	if err != nil {
		return nil, err
	}

	pub, err := data.ToBytes(agent.PublicKey)
	if err != nil {
		return nil, err
	}

	key, err := w.key(client.PrivateKey)
	if err != nil {
		return nil, err
	}

	mdata, err := messages.ClientOpen(sealed, pub, key)
	if err != nil {
		return nil, err
	}

	schema, err := statik.ReadFile(statik.EndpointJSONSchema)
	if err != nil {
		return nil, err
	}

	if !util.ValidateJSON(schema, mdata) {
		return nil, fmt.Errorf(
			"failed to validate endpoint for chan %s", ch.ID)
	}

	var msg ept.Message
	json.Unmarshal(mdata, &msg)

	return &msg, nil
}

// ClientPreEndpointMsgSOMCGet decodes endpoint message, saves it in the DB and
// triggers product configuration.
func (w *Worker) ClientPreEndpointMsgSOMCGet(job *data.Job) error {
	var sealed []byte
	if err := parseJobData(job, &sealed); err != nil {
		return err
	}

	var ch data.Channel
	err := data.FindByPrimaryKeyTo(w.db.Querier, &ch, job.RelatedID)
	if err != nil {
		return err
	}

	msg, err := w.decodeEndpoint(&ch, sealed)
	if err != nil {
		return err
	}

	var offer data.Offering
	err = data.FindByPrimaryKeyTo(w.db.Querier, &offer, ch.Offering)
	if err != nil {
		return err
	}

	params, _ := json.Marshal(msg.AdditionalParams)

	return w.db.InTransaction(func(tx *reform.TX) error {
		raddr := pointer.ToString(msg.PaymentReceiverAddress)
		saddr := pointer.ToString(msg.ServiceEndpointAddress)
		endp := data.Endpoint{
			ID:                     util.NewUUID(),
			Template:               offer.Template,
			Channel:                ch.ID,
			Hash:                   msg.TemplateHash,
			RawMsg:                 data.FromBytes(sealed),
			Status:                 data.MsgUnpublished,
			PaymentReceiverAddress: raddr,
			ServiceEndpointAddress: saddr,
			Username:               pointer.ToString(msg.Username),
			Password:               pointer.ToString(msg.Password),
			AdditionalParams:       params,
		}
		if err = w.db.Insert(&endp); err != nil {
			return err
		}

		return w.addJob(data.JobClientAfterEndpointMsgSOMCGet,
			data.JobEndpoint, endp.ID)
	})
}

// ClientAfterEndpointMsgSOMCGet cofigures a product.
func (w *Worker) ClientAfterEndpointMsgSOMCGet(job *data.Job) error {
	var endp data.Endpoint
	err := data.FindByPrimaryKeyTo(w.db.Querier, &endp, job.RelatedID)
	if err != nil {
		return err
	}

	if err := w.deployConfig(w.db, job.RelatedID, w.clientVPN.ConfigDir); err != nil {
		return err
	}

	return w.db.InTransaction(func(tx *reform.TX) error {
		endp.Status = data.MsgChPublished
		if err := data.Save(tx.Querier, &endp); err != nil {
			return err
		}

		var ch data.Channel
		err = data.FindByPrimaryKeyTo(w.db.Querier, &ch, endp.Channel)
		if err != nil {
			return err
		}

		ch.ServiceStatus = data.ServiceSuspended
		return data.Save(tx.Querier, &ch)
	})
}

// ClientAfterUncooperativeClose changed channel status
// to closed uncooperative.
func (w *Worker) ClientAfterUncooperativeClose(job *data.Job) error {
	ch, err := w.relatedChannel(job, data.JobClientAfterUncooperativeClose)
	if err != nil {
		return err
	}

	ch.ChannelStatus = data.ChannelClosedUncoop
	return data.Save(w.db.Querier, ch)
}
