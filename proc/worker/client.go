package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/client/svcrun"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/messages"
	"github.com/privatix/dappctrl/messages/ept"
	"github.com/privatix/dappctrl/messages/offer"
	"github.com/privatix/dappctrl/somc"
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

func (w *Worker) clientValidateChannelForClose(
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
		w.logger.Error("no supply for offering hash: %v",
			offerHash.Hex())
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

	auth := bind.NewKeyedTransactor(key)
	auth.GasLimit = w.gasConf.PSC.CreateChannel
	tx, err := w.ethBack.PSCCreateChannel(auth,
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

	var offering data.Offering
	err = data.FindByPrimaryKeyTo(w.db.Querier, &offering, jdata.Offering)
	if err != nil {
		return err
	}

	deposit, err := w.checkDeposit(&acc, &offering)
	if err != nil {
		return err
	}

	offerHash, err := data.ToHash(offering.Hash)
	if err != nil {
		return err
	}

	err = w.clientPreChannelCreateCheckSupply(&offering, offerHash)
	if err != nil {
		return err
	}

	msgRawBytes, err := data.ToBytes(offering.RawMsg)
	if err != nil {
		return err
	}

	msgRaw, _ := messages.UnpackSignature(msgRawBytes)

	msg := offer.Message{}
	if err := json.Unmarshal(msgRaw, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal offering msg: %v", err)
	}

	pubkB, err := data.ToBytes(msg.AgentPubKey)
	if err != nil {
		return fmt.Errorf("failed to decode agent pub key")
	}

	pubkey, err := crypto.UnmarshalPubkey(pubkB)
	if err != nil {
		return fmt.Errorf("failed to converts bytes to a secp256k1" +
			" public key")
	}

	agentEthAddr := data.FromBytes(crypto.PubkeyToAddress(*pubkey).Bytes())

	_, err = w.db.FindOneFrom(data.UserTable, "eth_addr", agentEthAddr)
	if err == sql.ErrNoRows {
		err = w.db.Insert(&data.User{
			ID:        util.NewUUID(),
			EthAddr:   agentEthAddr,
			PublicKey: msg.AgentPubKey,
		})
		if err != nil {
			return fmt.Errorf("failed to insert agent user rec: %v",
				err)
		}
	}

	return w.clientPreChannelCreateSaveTX(
		job, &acc, &offering, offerHash, deposit)
}

// ClientAfterChannelCreate activates channel and triggers endpoint retrieval.
func (w *Worker) ClientAfterChannelCreate(job *data.Job) error {
	var ch data.Channel
	err := data.FindByPrimaryKeyTo(w.db.Querier, &ch, job.RelatedID)
	if err != nil {
		return err
	}

	ethLog, err := w.ethLog(job)
	if err != nil {
		return err
	}

	ch.Block = uint32(ethLog.BlockNumber)
	ch.ChannelStatus = data.ChannelActive
	if err = data.Save(w.db.Querier, &ch); err != nil {
		return err
	}

	key, err := w.KeyFromChannelData(ch.ID)
	if err != nil {
		return err
	}

	endpointParams, err := w.somc.GetEndpoint(key)
	if err != nil {
		return fmt.Errorf("failed to get endpoint for chan %s: %s",
			ch.ID, err)
	}

	var ep *somc.EndpointParams
	if err := json.Unmarshal(endpointParams, &ep); err != nil {
		return err
	}

	err = w.addJobWithData(data.JobClientPreEndpointMsgSOMCGet,
		data.JobChannel, ch.ID, ep.Endpoint)
	if err != nil {
		return fmt.Errorf("failed to add "+
			"JobClientPreEndpointMsgSOMCGet job: %s", err)
	}

	client, err := w.account(ch.Client)
	if err != nil {
		return err
	}

	return w.addJob(
		data.JobAccountUpdateBalances, data.JobAccount, client.ID)
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
		changedTime := time.Now()
		ch.ServiceChangedTime = &changedTime
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
	if err := data.Save(w.db.Querier, ch); err != nil {
		return err
	}

	if ch.ServiceStatus != data.ServiceTerminated {
		_, err = w.processor.TerminateChannel(ch.ID, data.JobTask, false)
		if err != nil {
			return err
		}
	}

	agent, err := w.account(ch.Agent)
	if err != nil {
		return err
	}

	return w.addJob(
		data.JobAccountUpdateBalances, data.JobAccount, agent.ID)
}

// ClientAfterCooperativeClose changed channel status
// to closed cooperative and launches of terminate service procedure.
func (w *Worker) ClientAfterCooperativeClose(job *data.Job) error {
	ch, err := w.relatedChannel(job, data.JobClientAfterCooperativeClose)
	if err != nil {
		return err
	}

	ch.ChannelStatus = data.ChannelClosedCoop
	if err := data.Save(w.db.Querier, ch); err != nil {
		return err
	}

	if ch.ServiceStatus != data.ServiceTerminated {
		_, err = w.processor.TerminateChannel(ch.ID, data.JobTask, false)
		if err != nil {
			return err
		}
	}

	agent, err := w.account(ch.Agent)
	if err != nil {
		return err
	}

	return w.addJob(
		data.JobAccountUpdateBalances, data.JobAccount, agent.ID)
}

func (w *Worker) stopService(channel string) error {
	if err := w.runner.Stop(channel); err != nil {
		if err != svcrun.ErrNotRunning {
			return fmt.Errorf("failed to stop service: %s", err)
		}
		w.logger.Warn("failed to stop service: %s", err)
	}
	return nil
}

// ClientPreServiceTerminate terminates service.
func (w *Worker) ClientPreServiceTerminate(job *data.Job) error {
	ch, err := w.relatedChannel(job, data.JobClientPreServiceTerminate)
	if err != nil {
		return err
	}

	if err := w.stopService(ch.ID); err != nil {
		return err
	}

	ch.ServiceStatus = data.ServiceTerminated
	changedTime := time.Now()
	ch.ServiceChangedTime = &changedTime
	return data.Save(w.db.Querier, ch)
}

// ClientPreServiceSuspend suspends service.
func (w *Worker) ClientPreServiceSuspend(job *data.Job) error {
	ch, err := w.relatedChannel(job, data.JobClientPreServiceSuspend)
	if err != nil {
		return err
	}

	if err := w.stopService(ch.ID); err != nil {
		return err
	}

	ch.ServiceStatus = data.ServiceSuspended
	changedTime := time.Now()
	ch.ServiceChangedTime = &changedTime
	return data.Save(w.db.Querier, ch)
}

// ClientPreServiceUnsuspend activates service.
func (w *Worker) ClientPreServiceUnsuspend(job *data.Job) error {
	ch, err := w.relatedChannel(job, data.JobClientPreServiceUnsuspend)
	if err != nil {
		return err
	}

	if err := w.runner.Start(ch.ID); err != nil {
		return fmt.Errorf("failed to start service: %s", err)
	}

	ch.ServiceStatus = data.ServiceActive
	changedTime := time.Now()
	ch.ServiceChangedTime = &changedTime
	return data.Save(w.db.Querier, ch)
}

func (w *Worker) assertCanSettle(ctx context.Context, client,
	agent common.Address, block uint32,
	hash [common.HashLength]byte) error {
	_, _, settleBlock, _, err := w.ethBack.PSCGetChannelInfo(
		&bind.CallOpts{}, client, agent, block, hash)
	if err != nil {
		return err
	}

	curr, err := w.ethBack.LatestBlockNumber(ctx)
	if err != nil {
		return err
	}

	if curr.Uint64() < uint64(settleBlock) {
		return fmt.Errorf(
			"cannot settle yet (min. block %d, current %d)",
			uint64(settleBlock), curr.Uint64())
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

	err = w.assertCanSettle(ctx, client, agent, ch.Block, offerHash)
	if err != nil {
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

func (w *Worker) doClientPreUncooperativeCloseRequestAndSaveTx(job *data.Job,
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

// ClientPreUncooperativeCloseRequest requests the closing of the channel
// and starts the challenge period.
func (w *Worker) ClientPreUncooperativeCloseRequest(job *data.Job) error {
	ch, err := w.relatedChannel(job, data.JobClientPreUncooperativeCloseRequest)
	if err != nil {
		return err
	}

	jdata, err := w.publishData(job)
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

	if err := w.clientValidateChannelForClose(ch); err != nil {
		return err
	}

	return w.doClientPreUncooperativeCloseRequestAndSaveTx(job, ch, acc,
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

	blocks, err := data.ReadUintSetting(
		w.db.Querier, data.SettingEthChallengePeriod)
	if err != nil {
		return err
	}

	return w.addJobWithDelay(
		data.JobClientPreUncooperativeClose, data.JobChannel,
		ch.ID, time.Duration(blocks)*eth.BlockDuration)
}

// ClientAfterOfferingMsgBCPublish creates offering.
func (w *Worker) ClientAfterOfferingMsgBCPublish(job *data.Job) error {
	ethLog, err := w.ethLog(job)
	if err != nil {
		return err
	}

	logOfferingCreated, err := extractLogOfferingCreated(ethLog)
	if err != nil {
		return err
	}

	return w.clientRetrieveAndSaveOfferingFromSOMC(job, ethLog.BlockNumber,
		logOfferingCreated.agentAddr, logOfferingCreated.offeringHash)
}

// ClientAfterOfferingPopUp updates offering in db or retrieves from somc
// if new.
func (w *Worker) ClientAfterOfferingPopUp(job *data.Job) error {
	ethLog, err := w.ethLog(job)
	if err != nil {
		return err
	}

	logOfferingPopUp, err := extractLogOfferingPopUp(ethLog)
	if err != nil {
		return err
	}

	offering := data.Offering{}
	hash := data.FromBytes(logOfferingPopUp.offeringHash.Bytes())
	err = w.db.FindOneTo(&offering, "hash", hash)
	if err == sql.ErrNoRows {
		// New offering. Get from somc.
		return w.clientRetrieveAndSaveOfferingFromSOMC(job,
			ethLog.BlockNumber, logOfferingPopUp.agentAddr,
			logOfferingPopUp.offeringHash)
	}
	if err != nil {
		return err
	}

	// Existing offering, just update offering status.
	offering.BlockNumberUpdated = ethLog.BlockNumber

	return w.db.Update(&offering)
}

func (w *Worker) clientRetrieveAndSaveOfferingFromSOMC(
	job *data.Job, block uint64, agentAddr common.Address, hash common.Hash) error {
	offeringsData, err := w.somc.FindOfferings([]string{
		data.FromBytes(hash.Bytes())})
	if err != nil {
		return fmt.Errorf("failed to find offering: %v", err)
	}

	offering, err := w.fillOfferingFromSOMCReply(
		job.RelatedID, data.FromBytes(agentAddr.Bytes()),
		block, offeringsData)
	if err != nil {
		return fmt.Errorf("failed to fill offering: %v", err)
	}

	if err := w.db.Insert(offering); err != nil {
		return fmt.Errorf("failed to insert offering: %v", err)
	}

	return nil
}

func (w *Worker) fillOfferingFromSOMCReply(relID, agentAddr string, blockNumber uint64, offeringsData []somc.OfferingData) (*data.Offering, error) {
	if len(offeringsData) == 0 {
		return nil, fmt.Errorf("no offering returned from somc")
	}

	offeringData := offeringsData[0]
	if err := w.db.FindOneTo(&data.Offering{}, "hash", offeringData.Hash); err == nil {
		return nil, fmt.Errorf("offering exists with hash: %s", offeringData.Hash)
	}

	msgRaw, sig := messages.UnpackSignature(offeringData.Offering)

	msg := offer.Message{}
	if err := json.Unmarshal(msgRaw, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal offering data: %v", err)
	}

	pubk, err := data.ToBytes(msg.AgentPubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode agent pub key")
	}

	if !messages.VerifySignature(pubk, crypto.Keccak256(msgRaw), sig) {
		return nil, fmt.Errorf("wrong signature")
	}

	template, err := w.templateByHash(msg.TemplateHash)
	if err != nil {
		return nil, fmt.Errorf("could not find template by given hash: %v. got: %v",
			msg.TemplateHash, err)
	}

	product := &data.Product{}
	if err := w.db.FindOneTo(product, "offer_tpl_id", template.ID); err != nil {
		return nil, fmt.Errorf("could not find the product: %v", err)
	}

	return &data.Offering{
		ID:                 relID,
		Template:           template.ID,
		Product:            product.ID,
		Hash:               offeringData.Hash,
		Status:             data.MsgChPublished,
		OfferStatus:        data.OfferRegister,
		BlockNumberUpdated: blockNumber,
		Agent:              agentAddr,
		RawMsg:             data.FromBytes(offeringData.Offering),
		ServiceName:        product.Name,
		Country:            msg.Country,
		Supply:             msg.ServiceSupply,
		CurrentSupply:      msg.ServiceSupply,
		UnitName:           msg.UnitName,
		UnitType:           msg.UnitType,
		BillingType:        msg.BillingType,
		SetupPrice:         msg.SetupPrice,
		UnitPrice:          msg.UnitPrice,
		MinUnits:           msg.MinUnits,
		MaxUnit:            msg.MaxUnit,
		BillingInterval:    msg.BillingInterval,
		MaxBillingUnitLag:  msg.MaxBillingUnitLag,
		MaxSuspendTime:     msg.MaxSuspendTime,
		MaxInactiveTimeSec: msg.MaxInactiveTimeSec,
		FreeUnits:          msg.FreeUnits,
		AdditionalParams:   msg.ServiceSpecificParameters,
	}, nil
}

// ClientAfterOfferingDelete sets offer status to `remove`;
func (w *Worker) ClientAfterOfferingDelete(job *data.Job) error {
	return w.updateRelatedOffering(
		job, data.JobClientAfterOfferingDelete, data.OfferRemove)
}
