package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/client/somc"
	"github.com/privatix/dappctrl/country"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/messages"
	"github.com/privatix/dappctrl/messages/ept"
	"github.com/privatix/dappctrl/messages/offer"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

func (w *Worker) checkDeposit(logger log.Logger, acc *data.Account,
	offer *data.Offering, deposit uint64) error {
	logger = logger.Add("deposit", deposit)

	addr, err := data.HexToAddress(acc.EthAddr)
	if err != nil {
		logger.Error(err.Error())
		return ErrParseEthAddr
	}

	amount, err := w.ethBack.PSCBalanceOf(&bind.CallOpts{}, addr)
	if err != nil {
		logger.Error(err.Error())
		return ErrPSCReturnBalance
	}

	if deposit > amount {
		return ErrInsufficientPSCBalance
	}

	return nil
}

func (w *Worker) clientValidateChannelForClose(
	ch *data.Channel) error {
	// check channel status
	if ch.ChannelStatus != data.ChannelActive &&
		ch.ChannelStatus != data.ChannelPending {
		return ErrInvalidChannelStatus
	}

	// check service status
	if ch.ServiceStatus != data.ServiceTerminated &&
		ch.ServiceStatus != data.ServicePending {
		return ErrInvalidServiceStatus
	}

	// check receipt balance
	if ch.ReceiptBalance > ch.TotalDeposit {
		return ErrChannelReceiptBalance
	}

	return nil
}

func (w *Worker) clientPreChannelCreateCheckSupply(logger log.Logger,
	offer *data.Offering, offerHash common.Hash) error {
	_, _, _, supplyFromContract, _, _, err := w.ethBack.PSCGetOfferingInfo(
		&bind.CallOpts{}, offerHash)
	if err != nil {
		logger.Error(err.Error())
		return ErrPSCOfferingSupply
	}

	if supplyFromContract == 0 || offer.CurrentSupply == 0 {
		return ErrOfferingNoSupply
	}

	return nil
}

func (w *Worker) clientPreChannelCreateSaveTX(logger log.Logger,
	job *data.Job, acc *data.Account, offer *data.Offering,
	offerHash common.Hash, deposit uint64, gasPrice *big.Int) error {
	agentAddr, err := data.HexToAddress(offer.Agent)
	if err != nil {
		logger.Error(err.Error())
		return ErrParseEthAddr
	}

	key, err := w.key(logger, acc.PrivateKey)
	if err != nil {
		return err
	}

	auth := bind.NewKeyedTransactor(key)
	auth.GasLimit = w.gasConf.PSC.CreateChannel
	auth.GasPrice = gasPrice
	tx, err := w.ethBack.PSCCreateChannel(auth, agentAddr, offerHash, uint64(deposit))
	if err != nil {
		logger.Add("GasLimit", auth.GasLimit,
			"GasPrice", auth.GasPrice).Error(err.Error())
		return ErrPSCCreateChannel
	}

	if err := w.saveEthTX(logger, job, tx, "CreateChannel", data.JobChannel,
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
	err = data.Insert(w.db.Querier, &ch)
	if err != nil {
		logger.Add("channel", ch).Error(err.Error())
		return ErrInternal
	}

	return nil
}

// ClientPreChannelCreateData is a job data for ClientPreChannelCreate.
type ClientPreChannelCreateData struct {
	Account  string `json:"account"`
	Offering string `json:"offering"`
	GasPrice uint64 `json:"gasPrice"`
	Deposit  uint64 `json:"deposit"`
}

// ClientPreChannelCreate triggers a channel creation.
func (w *Worker) ClientPreChannelCreate(job *data.Job) error {
	logger := w.logger.Add("method", "ClientPreChannelCreate", "job", job)

	var jdata ClientPreChannelCreateData
	if err := w.unmarshalDataTo(logger, job.Data, &jdata); err != nil {
		return err
	}

	acc, err := w.accountByPK(logger, jdata.Account)
	if err != nil {
		return err
	}

	offering, err := w.offering(logger, jdata.Offering)
	if err != nil {
		return err
	}

	logger = logger.Add("account", acc, "offering", offering)

	deposit := jdata.Deposit
	if jdata.Deposit == 0 {
		deposit = data.ComputeDeposit(offering, offering.MinUnits)
	}

	if err := w.checkDeposit(logger, acc, offering, deposit); err != nil {
		return err
	}

	if deposit < data.ComputeDeposit(offering, offering.MinUnits) {
		return ErrSmallDeposit
	}

	offerHash, err := data.HexToHash(offering.Hash)
	if err != nil {
		logger.Error(err.Error())
		return ErrParseOfferingHash
	}

	err = w.clientPreChannelCreateCheckSupply(logger, offering, offerHash)
	if err != nil {
		return err
	}

	msgRawBytes, err := data.ToBytes(offering.RawMsg)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	msgRaw, _ := messages.UnpackSignature(msgRawBytes)

	msg := offer.Message{}
	if err := json.Unmarshal(msgRaw, &msg); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	logger = logger.Add("offeringMessage", msg)

	pubkB, err := data.ToBytes(msg.AgentPubKey)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	pubkey, err := crypto.UnmarshalPubkey(pubkB)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	agentEthAddr := data.HexFromBytes(crypto.PubkeyToAddress(*pubkey).Bytes())

	_, err = w.db.FindOneFrom(data.UserTable, "eth_addr", agentEthAddr)
	if err == sql.ErrNoRows {
		err = w.db.Insert(&data.User{
			ID:        util.NewUUID(),
			EthAddr:   agentEthAddr,
			PublicKey: msg.AgentPubKey,
		})
		if err != nil {
			logger.Error(err.Error())
			return ErrInternal
		}
	}

	var gasPrice *big.Int
	if jdata.GasPrice != 0 {
		gasPrice = new(big.Int).SetUint64(jdata.GasPrice)
	}

	return w.clientPreChannelCreateSaveTX(logger,
		job, acc, offering, offerHash, deposit, gasPrice)
}

// ClientAfterChannelCreate activates channel and triggers endpoint retrieval.
func (w *Worker) ClientAfterChannelCreate(job *data.Job) error {
	logger := w.logger.Add("method", "ClientAfterChannelCreate", "job", job)

	ch, err := w.relatedChannel(logger, job, data.JobClientAfterChannelCreate)
	if err != nil {
		return err
	}

	ethLog, err := w.ethLog(logger, job)
	if err != nil {
		return err
	}

	logger = logger.Add("channel", ch, "ethLog", ethLog)

	ch.Block = uint32(ethLog.Block)
	ch.ChannelStatus = data.ChannelActive
	if err = w.saveRecord(logger, w.db.Querier, ch); err != nil {
		return err
	}

	key, err := w.keyFromChannelData(logger, ch.ID)
	if err != nil {
		return err
	}

	logger = logger.Add("endpointKey", key)

	var endpointMsgSealed []byte

	offering, err := w.offering(logger, ch.Offering)
	if err != nil {
		return err
	}

	if err == nil {
		client, err := w.somcClientBuilder.NewClient(offering.SOMCType, offering.SOMCData)
		if err != nil {
			logger.Error(err.Error())
			return ErrGetEndpoint
		}
		rawMsg, err := client.Endpoint(key)
		if err != nil {
			logger.Error(err.Error())
			return ErrGetEndpoint
		}
		endpointMsgSealed, err = data.ToBytes(rawMsg)
		if err != nil {
			logger.Error(err.Error())
			return ErrGetEndpoint
		}
	}

	err = w.addJobWithData(logger, nil, data.JobClientEndpointGet,
		data.JobChannel, ch.ID, &data.JobEndpointCreateData{EndpointSealed: endpointMsgSealed})
	if err != nil {
		return err
	}

	client, err := w.account(logger, ch.Client)
	if err != nil {
		return err
	}

	return w.addJob(logger, nil,
		data.JobAccountUpdateBalances, data.JobAccount, client.ID)
}

func (w *Worker) extractEndpointMessage(logger log.Logger,
	ch *data.Channel, sealed []byte) (*ept.Message, error) {
	client, err := w.account(logger, ch.Client)
	if err != nil {
		return nil, err
	}

	agent, err := w.user(logger, ch.Agent)
	if err != nil {
		return nil, err
	}

	pub, err := data.ToBytes(agent.PublicKey)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	key, err := w.key(logger, client.PrivateKey)
	if err != nil {
		return nil, err
	}

	mdata, err := messages.ClientOpen(sealed, pub, key)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrDecryptEndpointMsg
	}

	var msg ept.Message
	err = json.Unmarshal(mdata, &msg)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	tpl, err := w.templateByHash(logger, msg.TemplateHash)
	if err != nil {
		return nil, err
	}

	if !util.ValidateJSON(tpl.Raw, mdata) {
		return nil, ErrInvalidEndpoint
	}

	return &msg, nil
}

// ClientEndpointGet decodes endpoint message, saves it in the DB and
// triggers product configuration.
func (w *Worker) ClientEndpointGet(job *data.Job) error {
	logger := w.logger.Add("method", "ClientEndpointGet",
		"job", job)

	ch, err := w.relatedChannel(logger, job,
		data.JobClientEndpointGet)
	if err != nil {
		return err
	}

	var jdata data.JobEndpointCreateData
	if err := w.unmarshalDataTo(logger, job.Data, &jdata); err != nil {
		return err
	}

	msg, err := w.extractEndpointMessage(logger, ch, jdata.EndpointSealed)
	if err != nil {
		return err
	}

	offer, err := w.offering(logger, ch.Offering)
	if err != nil {
		return err
	}

	url := strings.Replace(w.countryConfig.URLTemplate,
		"{{ip}}", msg.ServiceEndpointAddress, 1)

	var countryStatus string

	c, err := country.GetCountry(w.countryConfig.Timeout, url,
		w.countryConfig.Field)
	if err != nil || len(c) != 2 {
		countryStatus = data.CountryStatusUnknown
	} else if c == offer.Country {
		countryStatus = data.CountryStatusValid
	} else {
		countryStatus = data.CountryStatusInvalid
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
			RawMsg:                 data.FromBytes(jdata.EndpointSealed),
			PaymentReceiverAddress: raddr,
			ServiceEndpointAddress: saddr,
			Username:               pointer.ToString(msg.Username),
			Password:               pointer.ToString(msg.Password),
			AdditionalParams:       params,
			CountryStatus:          pointer.ToString(countryStatus),
		}
		if err = w.db.Insert(&endp); err != nil {
			logger.Add("endpoint", endp).Error(err.Error())
			return ErrInternal
		}

		ch.ServiceStatus = data.ServiceSuspended
		changedTime := time.Now()
		ch.ServiceChangedTime = &changedTime
		// TODO: Review flow with service_changed_time.
		ch.PreparedAt = changedTime
		err = w.saveRecord(logger, w.db.Querier, ch)
		if err != nil {
			logger.Error(err.Error())
			return ErrInternal
		}

		return nil
	})
}

// ClientAfterUncooperativeClose changed channel status
// to closed uncooperative.
func (w *Worker) ClientAfterUncooperativeClose(job *data.Job) error {
	logger := w.logger.Add("method", "ClientAfterUncooperativeClose",
		"job", job)
	ch, err := w.relatedChannel(logger, job, data.JobClientAfterUncooperativeClose)
	if err != nil {
		return err
	}

	logger = logger.Add("channel", ch)

	ch.ChannelStatus = data.ChannelClosedUncoop
	if err := w.saveRecord(logger, w.db.Querier, ch); err != nil {
		return err
	}

	client, err := w.account(logger, ch.Client)
	if err != nil {
		return err
	}

	return w.addJob(logger, nil,
		data.JobAccountUpdateBalances, data.JobAccount, client.ID)
}

// ClientAfterCooperativeClose changed channel status
// to closed cooperative and launches of terminate service procedure.
func (w *Worker) ClientAfterCooperativeClose(job *data.Job) error {
	logger := w.logger.Add("method", "ClientAfterCooperativeClose",
		"job", job)

	ch, err := w.relatedChannel(logger, job,
		data.JobClientAfterCooperativeClose)
	if err != nil {
		return err
	}

	logger = logger.Add("channel", ch)

	ch.ChannelStatus = data.ChannelClosedCoop
	if err := w.saveRecord(logger, w.db.Querier, ch); err != nil {
		return err
	}

	if ch.ServiceStatus != data.ServiceTerminated {
		_, err = w.processor.TerminateChannel(ch.ID, data.JobTask, false)
		if err != nil {
			logger.Error(err.Error())
			return ErrTerminateChannel
		}
	}

	client, err := w.account(logger, ch.Client)
	if err != nil {
		return err
	}

	return w.addJob(logger, nil,
		data.JobAccountUpdateBalances, data.JobAccount, client.ID)
}

// ClientPreServiceTerminate terminates service.
func (w *Worker) ClientPreServiceTerminate(job *data.Job) error {
	logger := w.logger.Add("method", "ClientPreServiceTerminate", "job", job)

	ch, err := w.relatedChannel(logger,
		job, data.JobClientPreServiceTerminate)
	if err != nil {
		return err
	}

	logger = logger.Add("channel", ch)

	if ch.ServiceStatus == data.ServiceActive {
		ch.ServiceStatus = data.ServiceTerminating
	} else {
		ch.ServiceStatus = data.ServiceTerminated
	}

	changedTime := time.Now()
	ch.ServiceChangedTime = &changedTime
	err = w.saveRecord(logger, w.db.Querier, ch)
	if err != nil {
		return err
	}
	return nil
}

// ClientPreServiceSuspend suspends service.
func (w *Worker) ClientPreServiceSuspend(job *data.Job) error {
	logger := w.logger.Add("method", "ClientPreServiceSuspend", "job", job)

	ch, err := w.relatedChannel(logger, job, data.JobClientPreServiceSuspend)
	if err != nil {
		return err
	}

	logger = logger.Add("channel", ch)

	ch.ServiceStatus = data.ServiceSuspending
	changedTime := time.Now()
	ch.ServiceChangedTime = &changedTime
	err = w.saveRecord(logger, w.db.Querier, ch)
	if err != nil {
		return err
	}
	return nil
}

// ClientPreServiceUnsuspend activates service.
func (w *Worker) ClientPreServiceUnsuspend(job *data.Job) error {
	logger := w.logger.Add("method", "ClientPreServiceUnsuspend",
		"job", job)

	ch, err := w.relatedChannel(logger, job, data.JobClientPreServiceUnsuspend)
	if err != nil {
		return err
	}

	logger = logger.Add("channel", ch)

	ch.ServiceStatus = data.ServiceActivating
	changedTime := time.Now()
	ch.ServiceChangedTime = &changedTime
	return w.saveRecord(logger, w.db.Querier, ch)
}

func (w *Worker) blocksTillChallangeEnd(ctx context.Context, logger log.Logger,
	client, agent common.Address, block uint32,
	hash [common.HashLength]byte) (int64, error) {
	_, settleBlock, _, err := w.ethBack.PSCGetChannelInfo(
		&bind.CallOpts{}, client, agent, block, hash)
	if err != nil {
		logger.Error(err.Error())
		return 0, ErrPSCGetChannelInfo
	}

	curr, err := w.ethBack.LatestBlockNumber(ctx)
	if err != nil {
		logger.Error(err.Error())
		return 0, ErrEthLatestBlockNumber
	}

	return int64(settleBlock) - curr.Int64(), nil
}

func (w *Worker) settle(ctx context.Context, logger log.Logger,
	acc *data.Account, agent common.Address, block uint32,
	hash [common.HashLength]byte) (*types.Transaction, error) {
	key, err := w.key(logger, acc.PrivateKey)
	if err != nil {
		return nil, err
	}

	opts := bind.NewKeyedTransactor(key)
	opts.Context = ctx

	tx, err := w.ethBack.PSCSettle(opts, agent, block, hash)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrPSCSettle
	}

	return tx, nil
}

// ClientPreUncooperativeClose waiting for until the challenge
// period is over. Then deletes the channel and settles
// by transferring the balance to the Agent and the rest
// of the deposit back to the Client.
func (w *Worker) ClientPreUncooperativeClose(job *data.Job) error {
	logger := w.logger.Add("method", "ClientPreUncooperativeClose",
		"job", job)

	// If cooperative close was created for this channel, skip this job.
	err := w.db.SelectOneTo(&data.Job{},
		"WHERE related_id=$1 AND related_type=$2 AND type=$3",
		job.RelatedID, job.RelatedType, data.JobClientAfterCooperativeClose)
	if err == nil {
		job.Status = data.JobCanceled
		w.db.Save(job)
		return nil
	}

	ch, err := w.relatedChannel(logger, job,
		data.JobClientPreUncooperativeClose)
	if err != nil {
		return err
	}

	if ch.ChannelStatus == data.ChannelClosedCoop {
		logger.Warn("channel closed cooperatively")
		return nil
	}

	logger = logger.Add("channel", ch)

	agent, err := data.HexToAddress(ch.Agent)
	if err != nil {
		logger.Error(err.Error())
		return ErrParseEthAddr
	}

	client, err := data.HexToAddress(ch.Client)
	if err != nil {
		logger.Error(err.Error())
		return ErrParseEthAddr
	}

	acc, err := w.account(logger, ch.Client)
	if err != nil {
		return err
	}

	offer, err := w.offering(logger, ch.Offering)
	if err != nil {
		return err
	}

	offerHash, err := data.HexToHash(offer.Hash)
	if err != nil {
		logger.Error(err.Error())
		return ErrParseOfferingHash
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	blocks, err := w.blocksTillChallangeEnd(ctx, logger, client, agent, ch.Block, offerHash)
	if err != nil {
		return err
	}
	if blocks > 0 {
		logger.Add("blocksTillChallangeEnd", blocks).Warn("in challange period")
		return ErrChallengePeriodIsNotOver
	}

	tx, err := w.settle(ctx, logger, acc, agent, ch.Block, offerHash)
	if err != nil {
		return err
	}

	if err := w.saveEthTX(logger, job, tx, "Settle",
		data.JobChannel, ch.ID, acc.EthAddr,
		data.HexFromBytes(w.pscAddr.Bytes())); err != nil {
		return err
	}

	ch.ChannelStatus = data.ChannelWaitUncoop

	return w.saveRecord(logger, w.db.Querier, ch)
}

// ClientPreChannelTopUpData is a job data for ClientPreChannelTopUp.
type ClientPreChannelTopUpData struct {
	Channel  string `json:"channel"`
	GasPrice uint64 `json:"gasPrice"`
}

func (w *Worker) clientPreChannelTopUpSaveTx(logger log.Logger, job *data.Job,
	ch *data.Channel, acc *data.Account, offer *data.Offering,
	gasPrice uint64, deposit uint64) error {
	agent, err := data.HexToAddress(ch.Agent)
	if err != nil {
		logger.Error(err.Error())
		return ErrParseOfferingHash
	}

	offerHash, err := data.HexToHash(offer.Hash)
	if err != nil {
		logger.Error(err.Error())
		return ErrParseOfferingHash
	}

	key, err := w.key(logger, acc.PrivateKey)
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
		logger.Add("GasLimit", opts.GasLimit,
			"GasPrice", opts.GasPrice).Error(err.Error())
		return ErrPSCTopUpChannel
	}

	return w.saveEthTX(logger, job, tx, "TopUpChannel",
		data.JobChannel, ch.ID, acc.EthAddr,
		data.HexFromBytes(w.pscAddr.Bytes()))
}

// ClientPreChannelTopUp checks client balance and creates transaction
// for increase the channel deposit.
func (w *Worker) ClientPreChannelTopUp(job *data.Job) error {
	logger := w.logger.Add("method", "ClientPreChannelTopUp", "job", job)

	ch, err := w.relatedChannel(logger, job, data.JobClientPreChannelTopUp)
	if err != nil {
		return err
	}

	offer, err := w.offering(logger, ch.Offering)
	if err != nil {
		return err
	}

	acc, err := w.account(logger, ch.Client)
	if err != nil {
		return err
	}

	logger = logger.Add("channel", ch, "offering", offer)

	var jdata data.JobTopUpChannelData
	if err := w.unmarshalDataTo(logger, job.Data, &jdata); err != nil {
		return err
	}

	if err := w.checkDeposit(logger, acc, offer, jdata.Deposit); err != nil {
		return err
	}

	return w.clientPreChannelTopUpSaveTx(logger, job, ch, acc, offer,
		jdata.GasPrice, uint64(jdata.Deposit))
}

// ClientAfterChannelTopUp updates deposit of a channel.
func (w *Worker) ClientAfterChannelTopUp(job *data.Job) error {
	return w.afterChannelTopUp(job, data.JobClientAfterChannelTopUp)
}

func (w *Worker) doClientPreUncooperativeCloseRequestAndSaveTx(logger log.Logger,
	job *data.Job, ch *data.Channel, acc *data.Account, offer *data.Offering,
	gasPrice uint64) error {
	agent, err := data.HexToAddress(ch.Agent)
	if err != nil {
		logger.Error(err.Error())
		return ErrParseEthAddr
	}

	key, err := w.key(logger, acc.PrivateKey)
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

	offerHash, err := data.HexToHash(offer.Hash)
	if err != nil {
		logger.Error(err.Error())
		return ErrParseOfferingHash
	}

	if w.gasConf.PSC.UncooperativeClose != 0 {
		opts.GasLimit = w.gasConf.PSC.UncooperativeClose
	}

	tx, err := w.ethBack.PSCUncooperativeClose(opts, agent, ch.Block,
		offerHash, uint64(ch.ReceiptBalance))
	if err != nil {
		logger.Error(err.Error())
		return ErrPSCUncooperativeClose
	}

	if err := w.saveEthTX(logger, job, tx, "UncooperativeClose",
		data.JobChannel, ch.ID, acc.EthAddr,
		data.HexFromBytes(w.pscAddr.Bytes())); err != nil {
		return err
	}

	ch.ChannelStatus = data.ChannelWaitChallenge

	return w.saveRecord(logger, w.db.Querier, ch)
}

// ClientPreUncooperativeCloseRequest requests the closing of the channel
// and starts the challenge period.
func (w *Worker) ClientPreUncooperativeCloseRequest(job *data.Job) error {
	logger := w.logger.Add("method", "ClientPreUncooperativeCloseRequest",
		"job", job)

	ch, err := w.relatedChannel(logger, job,
		data.JobClientPreUncooperativeCloseRequest)
	if err != nil {
		return err
	}

	jdata, err := w.publishData(logger, job)
	if err != nil {
		return err
	}

	offer, err := w.offering(logger, ch.Offering)
	if err != nil {
		return err
	}

	acc, err := w.account(logger, ch.Client)
	if err != nil {
		return err
	}

	if err := w.clientValidateChannelForClose(ch); err != nil {
		return err
	}

	return w.doClientPreUncooperativeCloseRequestAndSaveTx(logger, job, ch,
		acc, offer, jdata.GasPrice)
}

// ClientAfterUncooperativeCloseRequest waits for the channel
// to uncooperative close, starts the service termination process.
func (w *Worker) ClientAfterUncooperativeCloseRequest(job *data.Job) error {
	logger := w.logger.Add("method", "ClientAfterUncooperativeCloseRequest",
		"job", job)

	ch, err := w.relatedChannel(logger, job,
		data.JobClientAfterUncooperativeCloseRequest)
	if err != nil {
		return err
	}

	ch.ChannelStatus = data.ChannelInChallenge
	if err = w.db.Update(ch); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	challengePeriod, err := data.ReadUintSetting(w.db.Querier,
		data.SettingsPeriodChallange)
	if err != nil {
		return err
	}

	return w.addJobWithDelay(logger, nil,
		data.JobClientPreUncooperativeClose, data.JobChannel,
		ch.ID, time.Duration(challengePeriod)*eth.BlockDuration)
}

// ClientAfterOfferingMsgBCPublish creates offering.
func (w *Worker) ClientAfterOfferingMsgBCPublish(job *data.Job) error {
	logger := w.logger.Add("method", "ClientAfterOfferingMsgBCPublish",
		"job", job)

	ethLog, err := w.ethLog(logger, job)
	if err != nil {
		return err
	}

	logOfferingCreated, err := extractLogOfferingCreated(logger, ethLog)
	if err != nil {
		return err
	}

	return w.clientRetrieveAndSaveOffering(logger, job,
		ethLog.Block, logOfferingCreated.somcType,
		logOfferingCreated.somcData, logOfferingCreated.agentAddr,
		logOfferingCreated.offeringHash, logOfferingCreated.currentSupply)
}

// ClientAfterOfferingPopUp updates offering in db or retrieves from somc
// if new.
func (w *Worker) ClientAfterOfferingPopUp(job *data.Job) error {
	logger := w.logger.Add("method", "ClientAfterOfferingPopUp", "job", job)

	ethLog, err := w.ethLog(logger, job)
	if err != nil {
		return err
	}

	logger = logger.Add("ethLog", ethLog)

	logOfferingPopUp, err := extractLogOfferingPopUp(logger, ethLog)
	if err != nil {
		return err
	}

	offering := data.Offering{}
	hash := data.HexFromBytes(logOfferingPopUp.offeringHash.Bytes())
	err = w.db.FindOneTo(&offering, "hash", hash)
	if err == sql.ErrNoRows {
		// New offering.
		return w.clientRetrieveAndSaveOffering(logger, job,
			ethLog.Block, logOfferingPopUp.somcType,
			logOfferingPopUp.somcData, logOfferingPopUp.agentAddr,
			logOfferingPopUp.offeringHash, logOfferingPopUp.currentSupply)
	}
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	// Existing offering, just update offering status.
	offering.BlockNumberUpdated = ethLog.Block
	offering.Status = data.OfferPoppedUp

	return w.saveRecord(logger, w.db.Querier, &offering)
}

func (w *Worker) clientRetrieveAndSaveOffering(logger log.Logger,
	job *data.Job, block uint64, somcType uint8, somcData data.Base64String,
	agentAddr common.Address, hash common.Hash, currentSupply uint16) error {
	logger = logger.Add("job", fmt.Sprintf("%+v", job), "block", block, "somcType",
		somcType, "somcData", somcData, "agentAddr", agentAddr, "hash", hash,
		"currentSupply", currentSupply)

	client, err := w.somcClientBuilder.NewClient(somcType, somcData)
	if err == somc.ErrUnknownSOMCType {
		logger.Warn(err.Error())
		return nil
	}
	if err != nil {
		logger.Error(err.Error())
		return ErrFetchOffering
	}
	hashHex := data.HexFromBytes(hash.Bytes())
	offeringRawMsg, err := client.Offering(hashHex)
	if err != nil {
		// Ignoring errors if did not receive an offering from an agent.
		logger.Add("offeringHash", hashHex).Warn(
			"could not fetch offerings, error: " + err.Error())
		return ErrFetchOffering
	}
	offeringRawMsgBytes, err := data.ToBytes(offeringRawMsg)
	if err != nil {
		logger.Error(err.Error())
		return ErrFetchOffering
	}
	offering, err := w.fillOfferingFromMsg(logger, offeringRawMsgBytes,
		block, data.HexFromBytes(agentAddr.Bytes()),
		data.HexFromBytes(hash.Bytes()), job.RelatedID,
		somcType, somcData)
	if err != nil {
		if err == ErrTemplateByHashNotFound {
			job.Status = data.JobCanceled
			w.db.Save(job)
			return nil
		}
		// Ignore all other errors except internal.
		if err != ErrInternal {
			logger.Warn(err.Error())
			return nil
		}
		return ErrFetchOffering
	}

	_, minDeposit, mSupply, _, _, _, err := w.ethBack.PSCGetOfferingInfo(
		&bind.CallOpts{}, hash)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	if minDeposit != data.ComputeDeposit(offering, offering.MinUnits) {
		return ErrOfferingDeposit
	}

	offering.Supply = mSupply
	offering.CurrentSupply = currentSupply

	if err := data.Insert(w.db.Querier, offering); err != nil {
		logger.Add("offering", offering).Error(err.Error())
		return ErrInternal
	}

	return nil
}

func (w *Worker) fillOfferingFromMsg(logger log.Logger, offering []byte,
	blockNumber uint64, agent, hash data.HexString, relID string,
	somcType uint8, somcData data.Base64String) (*data.Offering, error) {
	logger = logger.Add("offering", offering)
	_, err := w.offeringByHashString(logger, hash)
	if err == nil {
		return nil, ErrOfferingExists
	}

	hashBytes := common.BytesToHash(crypto.Keccak256(offering))

	// Check hash match to that in registered in blockchain.
	_, _, _, _, _, active, err := w.ethBack.PSCGetOfferingInfo(
		&bind.CallOpts{}, hashBytes)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	if !active {
		return nil, ErrOfferingNotActive
	}

	msgRaw, sig := messages.UnpackSignature(offering)

	msg := offer.Message{}
	if err := json.Unmarshal(msgRaw, &msg); err != nil {
		logger.Add("msgRaw", msgRaw).Error(err.Error())
		return nil, ErrInternal
	}

	logger = logger.Add("msg", fmt.Sprintf("%+v", msg))

	pubk, err := data.ToBytes(msg.AgentPubKey)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	if !messages.VerifySignature(pubk, crypto.Keccak256(msgRaw), sig) {
		return nil, ErrWrongOfferingMsgSignature
	}

	template, err := w.templateByHash(logger, msg.TemplateHash)
	if err != nil {
		return nil, err
	}

	// Validate offering JSON compliant with offering template JSON
	if !offer.ValidMsg(template.Raw, msg) {
		return nil, ErrOfferNotCorrespondToTemplate
	}

	product := &data.Product{}
	if err := w.db.SelectOneTo(
		product, "WHERE offer_tpl_id = $1 AND NOT is_server",
		template.ID); err != nil {
		logger.Error(err.Error())
		return nil, ErrProductNotFound
	}

	return &data.Offering{
		ID:                 relID,
		Template:           template.ID,
		Product:            product.ID,
		Hash:               hash,
		Status:             data.OfferRegistered,
		BlockNumberUpdated: blockNumber,
		Agent:              agent,
		RawMsg:             data.FromBytes(offering),
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
		SOMCType:           somcType,
		SOMCData:           somcData,
	}, nil
}

// ClientAfterOfferingDelete sets offer status to `remove`;
func (w *Worker) ClientAfterOfferingDelete(job *data.Job) error {
	return w.updateRelatedOffering(
		job, data.JobClientAfterOfferingDelete, data.OfferRemoved)
}

// DecrementCurrentSupply finds offering and decrements its current supply.
func (w *Worker) DecrementCurrentSupply(job *data.Job) error {
	logger := w.logger.Add("method", "DecrementCurrentSupply", "job", job)
	offering, err := w.relatedOffering(logger, job, data.JobDecrementCurrentSupply)
	if err != nil {
		return err
	}

	offering.CurrentSupply--

	err = data.Save(w.db.Querier, offering)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return nil
}

// IncrementCurrentSupply finds offering and increments its current supply.
func (w *Worker) IncrementCurrentSupply(job *data.Job) error {
	logger := w.logger.Add("method", "IncrementCurrentSupply", "job", job)
	offering, err := w.relatedOffering(logger, job,
		data.JobIncrementCurrentSupply)
	if err != nil {
		return err
	}

	offering.CurrentSupply++

	err = data.Save(w.db.Querier, offering)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return nil
}
