package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/messages"
	"github.com/privatix/dappctrl/messages/ept"
	"github.com/privatix/dappctrl/statik"
	"github.com/privatix/dappctrl/util"
)

func (w *Worker) checkDeposit(acc *data.Account,
	offer *data.Offering) (uint64, error) {
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

func (w *Worker) clientPreUncooperativeCloseRequestCheckChannel(
	ch *data.Channel) error {
	// check channel status
	if ch.ChannelStatus != data.ChannelActive &&
		ch.ChannelStatus != data.ChannelPending {
		return ErrInvalidChStatus
	}

	// check service status
	if ch.ServiceStatus == data.ServiceTerminated {
		return ErrInvalidServiceStatus
	}

	// check receipt balance
	if ch.ReceiptBalance > ch.TotalDeposit {
		return ErrChReceiptBalance
	}

	return nil
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
	Offering string `json:"offering"`
	GasPrice uint64 `json:"gasPrice"`
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
	err = data.FindByPrimaryKeyTo(w.db.Querier, &offer, jdata.Offering)
	if err != nil {
		return err
	}

	deposit, err := w.checkDeposit(&acc, &offer)
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

func (w *Worker) waitSettlePeriod(ctx context.Context, client,
	agent common.Address, block uint32,
	hash [common.HashLength]byte) error {
	_, _, settleBlock, _, err := w.ethBack.PSCGetChannelInfo(
		&bind.CallOpts{}, client, agent, block, hash)
	if err != nil {
		return err
	}

	for {
		block, err := w.ethBack.LatestBlockNumber(ctx)
		if err != nil {
			return err
		}

		if block.Uint64() >= uint64(settleBlock) {
			break
		}

		// TODO(maxim) hardcoded timeout
		time.Sleep(time.Second * 10)
	}
	return nil
}

func (w *Worker) settle(ctx context.Context, acc *data.Account,
	agent common.Address, block uint32,
	hash [common.HashLength]byte) (*types.Transaction, error) {
	key, err := w.key(acc.PrivateKey)
	if err != nil {
		return nil, err
	}

	opts := bind.NewKeyedTransactor(key)
	opts.Context = ctx

	return w.ethBack.PSCSettle(opts, agent, block, hash)
}

// ClientPreUncooperativeClose waiting for until the challenge
// period is over. Then deletes the channel and settles
// by transferring the balance to the Agent and the rest
// of the deposit back to the Client.
func (w *Worker) ClientPreUncooperativeClose(job *data.Job) error {
	ch, err := w.relatedChannel(job,
		data.JobClientPreUncooperativeClose)
	if err != nil {
		return err
	}

	agent, err := data.ToAddress(ch.Agent)
	if err != nil {
		return err
	}

	client, err := data.ToAddress(ch.Client)
	if err != nil {
		return err
	}

	var acc data.Account
	if err := data.FindOneTo(w.db.Querier, &acc,
		"eth_addr", ch.Client); err != nil {
		return err
	}

	var offer data.Offering
	if err := data.FindByPrimaryKeyTo(w.db.Querier, &offer,
		ch.Offering); err != nil {
		return err
	}

	offerHash, err := data.ToHash(offer.Hash)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.waitSettlePeriod(ctx, client, agent, ch.Block,
		offerHash); err != nil {
		return err
	}

	tx, err := w.settle(ctx, &acc, agent, ch.Block, offerHash)
	if err != nil {
		return err
	}

	if err := w.saveEthTX(job, tx, "Settle",
		data.JobChannel, ch.ID, acc.EthAddr,
		data.FromBytes(w.pscAddr.Bytes())); err != nil {
		return err
	}

	ch.ChannelStatus = data.ChannelWaitUncoop

	return data.Save(w.db.Querier, ch)
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

// ClientAfterCooperativeClose changed channel status
// to closed cooperative and launches of terminate service procedure.
func (w *Worker) ClientAfterCooperativeClose(job *data.Job) error {
	var ch data.Channel
	if err := data.FindByPrimaryKeyTo(w.db.Querier, &ch,
		job.RelatedID); err != nil {
		return err
	}

	ch.ChannelStatus = data.ChannelClosedCoop
	if err := data.Save(w.db.Querier, &ch); err != nil {
		return err
	}

	return w.addJob(data.JobClientPreServiceTerminate,
		data.JobChannel, ch.ID)
}

// ClientPreChannelTopUpData is a job data for ClientPreChannelTopUp.
type ClientPreChannelTopUpData struct {
	Channel  string `json:"channel"`
	GasPrice uint64 `json:"gasPrice"`
}

func (w *Worker) clientPreChannelTopUpSaveTx(job *data.Job, ch *data.Channel,
	acc *data.Account, offer *data.Offering, gasPrice uint64,
	deposit *big.Int) error {
	agent, err := data.ToAddress(ch.Agent)
	if err != nil {
		return err
	}

	offerHash, err := data.ToHash(offer.Hash)
	if err != nil {
		return err
	}

	key, err := w.key(acc.PrivateKey)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := bind.NewKeyedTransactor(key)
	opts.Context = ctx
	if gasPrice != 0 {
		opts.GasPrice = new(big.Int).SetUint64(gasPrice)
	}
	if w.gasConf.PSC.TopUp != 0 {
		opts.GasLimit = w.gasConf.PSC.TopUp
	}

	tx, err := w.ethBack.PSCTopUpChannel(opts, agent, ch.Block,
		offerHash, deposit)
	if err != nil {
		return err
	}

	return w.saveEthTX(job, tx, "TopUpChannel",
		data.JobChannel, ch.ID, acc.EthAddr,
		data.FromBytes(w.pscAddr.Bytes()))
}

// ClientPreChannelTopUp checks client balance and creates transaction
// for increase the channel deposit.
func (w *Worker) ClientPreChannelTopUp(job *data.Job) error {
	var jdata ClientPreChannelTopUpData
	if err := parseJobData(job, &jdata); err != nil {
		return err
	}

	ch, err := w.channel(jdata.Channel)
	if err != nil {
		return err
	}

	offer, err := w.offering(ch.Offering)
	if err != nil {
		return err
	}

	acc, err := w.account(ch.Client)
	if err != nil {
		return err
	}

	deposit, err := w.checkDeposit(acc, offer)
	if err != nil {
		return err
	}

	return w.clientPreChannelTopUpSaveTx(job, ch, acc, offer,
		jdata.GasPrice, new(big.Int).SetUint64(deposit))
}

// ClientAfterChannelTopUp updates deposit of a channel.
func (w *Worker) ClientAfterChannelTopUp(job *data.Job) error {
	return w.afterChannelTopUp(job, data.JobClientAfterChannelTopUp)
}

func (w *Worker) clientPreUncooperativeCloseRequestSaveTx(job *data.Job,
	ch *data.Channel, acc *data.Account, offer *data.Offering,
	gasPrice uint64) error {
	agent, err := data.ToAddress(ch.Agent)
	if err != nil {
		return err
	}

	key, err := w.key(acc.PrivateKey)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := bind.NewKeyedTransactor(key)
	opts.Context = ctx
	if gasPrice != 0 {
		opts.GasPrice = new(big.Int).SetUint64(gasPrice)
	}

	offerHash, err := data.ToHash(offer.Hash)
	if err != nil {
		return err
	}

	if w.gasConf.PSC.UncooperativeClose != 0 {
		opts.GasLimit = w.gasConf.PSC.UncooperativeClose
	}

	tx, err := w.ethBack.PSCUncooperativeClose(opts, agent, ch.Block,
		offerHash, new(big.Int).SetUint64(ch.ReceiptBalance))
	if err != nil {
		return err
	}

	if err := w.saveEthTX(job, tx, "UncooperativeClose",
		data.JobChannel, ch.ID, acc.EthAddr,
		data.FromBytes(w.pscAddr.Bytes())); err != nil {
		return err
	}

	ch.ChannelStatus = data.ChannelWaitChallenge

	return data.Save(w.db.Querier, ch)
}

// ClientPreUncooperativeCloseRequestData is a job data
// for ClientPreUncooperativeCloseRequest.
type ClientPreUncooperativeCloseRequestData struct {
	Channel  string `json:"channel"`
	GasPrice uint64 `json:"gasPrice"`
}

// ClientPreUncooperativeCloseRequest requests the closing of the channel
// and starts the challenge period.
func (w *Worker) ClientPreUncooperativeCloseRequest(job *data.Job) error {
	var jdata ClientPreUncooperativeCloseRequestData
	if err := parseJobData(job, &jdata); err != nil {
		return err
	}

	ch, err := w.channel(jdata.Channel)
	if err != nil {
		return err
	}

	offer, err := w.offering(ch.Offering)
	if err != nil {
		return err
	}

	acc, err := w.account(ch.Client)
	if err != nil {
		return err
	}

	if err := w.clientPreUncooperativeCloseRequestCheckChannel(
		ch); err != nil {
		return err
	}

	return w.clientPreUncooperativeCloseRequestSaveTx(job, ch, acc,
		offer, jdata.GasPrice)
}

// ClientAfterUncooperativeCloseRequest waits for the channel
// to uncooperative close, starts the service termination process.
func (w *Worker) ClientAfterUncooperativeCloseRequest(job *data.Job) error {
	ch, err := w.relatedChannel(job,
		data.JobClientAfterUncooperativeCloseRequest)
	if err != nil {
		return err
	}

	ch.ChannelStatus = data.ChannelInChallenge
	if err = w.db.Update(ch); err != nil {
		return fmt.Errorf("could not update channel's"+
			" status: %v", err)
	}

	if err := w.addJob(data.JobClientPreUncooperativeClose,
		data.JobChannel, ch.ID); err != nil {
		return err
	}

	return w.addJob(data.JobClientPreServiceTerminate,
		data.JobChannel, ch.ID)
}

// ClientPreServiceTerminate terminates service.
func (w *Worker) ClientPreServiceTerminate(job *data.Job) error {
	return nil
}
