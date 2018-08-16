package worker

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/messages"
	"github.com/privatix/dappctrl/messages/offer"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

// AgentAfterChannelCreate registers client and creates pre service create job.
func (w *Worker) AgentAfterChannelCreate(job *data.Job) error {
	if w.isJobInvalid(job, data.JobAgentAfterChannelCreate, data.JobChannel) {
		return ErrInvalidJob
	}

	logger := w.logger.Add("method", "AgentAfterChannelCreate", "job", job)

	ethLog, err := w.ethLog(logger, job)
	if err != nil {
		return err
	}

	ethLogTx, err := w.ethLogTx(logger, ethLog)
	if err != nil {
		return err
	}

	client, newUser, err := w.newUser(logger, ethLogTx)
	if err != nil {
		return err
	}

	logger = logger.Add("client", client)

	tx, err := w.db.Begin()
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}
	defer tx.Rollback()

	if newUser {
		if err := tx.Insert(client); err != nil {
			logger.Error(err.Error())
			return ErrInternal
		}
	}

	logChannelCreated, err := extractLogChannelCreated(logger, ethLog)
	if err != nil {
		return err
	}

	offering, err := w.offeringByHash(logger, logChannelCreated.offeringHash)
	if err != nil {
		return err
	}

	offering.CurrentSupply--
	if err := tx.Update(offering); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	channel := &data.Channel{
		ID:            job.RelatedID,
		Client:        data.HexFromBytes(logChannelCreated.clientAddr.Bytes()),
		Agent:         data.HexFromBytes(logChannelCreated.agentAddr.Bytes()),
		TotalDeposit:  logChannelCreated.deposit.Uint64(),
		ChannelStatus: data.ChannelActive,
		ServiceStatus: data.ServicePending,
		Offering:      offering.ID,
		Block:         uint32(ethLog.BlockNumber),
	}

	if err := tx.Insert(channel); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("unable to commit changes: %v", err)
	}

	return w.addJob(logger, data.JobAgentPreEndpointMsgCreate,
		data.JobChannel, channel.ID)
}

// AgentAfterChannelTopUp updates deposit of a channel.
func (w *Worker) AgentAfterChannelTopUp(job *data.Job) error {
	return w.afterChannelTopUp(job, data.JobAgentAfterChannelTopUp)
}

// AgentAfterUncooperativeCloseRequest sets channel's status to challenge period.
func (w *Worker) AgentAfterUncooperativeCloseRequest(job *data.Job) error {
	logger := w.logger.Add("method", "AgentAfterUncooperativeCloseRequest",
		"job", job)
	channel, err := w.relatedChannel(logger, job,
		data.JobAgentAfterUncooperativeCloseRequest)
	if err != nil {
		return err
	}

	logger = logger.Add("channel", channel)

	if channel.ServiceStatus != data.ServiceTerminated {
		_, err = w.processor.TerminateChannel(
			channel.ID, data.JobTask, true)
		if err != nil {
			logger.Error(err.Error())
			return ErrTerminateChannel
		}
	}

	channel.ChannelStatus = data.ChannelInChallenge
	if err = w.db.Update(channel); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return nil
}

func (w *Worker) incrementCurrentSupply(logger log.Logger, pk string) error {
	offering, err := w.offering(logger, pk)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	offering.CurrentSupply++
	if err := w.db.Update(offering); err != nil {
		logger.Add("offering", offering).Error(err.Error())
		return ErrInternal
	}

	return nil
}

// AgentAfterUncooperativeClose marks channel closed uncoop.
func (w *Worker) AgentAfterUncooperativeClose(job *data.Job) error {
	logger := w.logger.Add("method", "AgentAfterUncooperativeClose",
		"job", job)

	channel, err := w.relatedChannel(logger, job,
		data.JobAgentAfterUncooperativeClose)
	if err != nil {
		return err
	}

	if channel.ServiceStatus != data.ServiceTerminated {
		_, err = w.processor.TerminateChannel(
			channel.ID, data.JobTask, true)
		if err != nil {
			logger.Error(err.Error())
			return ErrTerminateChannel
		}
	}

	channel.ChannelStatus = data.ChannelClosedUncoop
	if err = w.db.Update(channel); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	if err := w.incrementCurrentSupply(logger, channel.Offering); err != nil {
		return err
	}

	agent, err := w.account(logger, channel.Agent)
	if err != nil {
		return err
	}

	return w.addJob(logger,
		data.JobAccountUpdateBalances, data.JobAccount, agent.ID)
}

// AgentAfterCooperativeClose marks channel as closed coop.
func (w *Worker) AgentAfterCooperativeClose(job *data.Job) error {
	logger := w.logger.Add("method", "AgentAfterCooperativeClose",
		"job", job)
	channel, err := w.relatedChannel(logger,
		job, data.JobAgentAfterCooperativeClose)
	if err != nil {
		return err
	}

	channel.ChannelStatus = data.ChannelClosedCoop
	if err := w.db.Update(channel); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	if err := w.incrementCurrentSupply(logger, channel.Offering); err != nil {
		return err
	}

	agent, err := w.account(logger, channel.Agent)
	if err != nil {
		return err
	}

	return w.addJob(logger,
		data.JobAccountUpdateBalances, data.JobAccount, agent.ID)
}

// AgentPreServiceSuspend marks service as suspended.
func (w *Worker) AgentPreServiceSuspend(job *data.Job) error {
	logger := w.logger.Add("method", "AgentPreServiceSuspend", "job", job)
	_, err := w.agentUpdateServiceStatus(logger, job,
		data.JobAgentPreServiceSuspend)
	return err
}

// AgentPreServiceUnsuspend marks service as active.
func (w *Worker) AgentPreServiceUnsuspend(job *data.Job) error {
	logger := w.logger.Add("method", "AgentPreServiceSuspend", "job", job)
	_, err := w.agentUpdateServiceStatus(logger, job,
		data.JobAgentPreServiceUnsuspend)
	return err
}

// AgentPreServiceTerminate terminates the service.
func (w *Worker) AgentPreServiceTerminate(job *data.Job) error {
	logger := w.logger.Add("method", "AgentPreServiceSuspend", "job", job)
	channel, err := w.agentUpdateServiceStatus(logger, job,
		data.JobAgentPreServiceTerminate)
	if err != nil {
		return err
	}
	return w.agentCooperativeClose(logger, job, channel)
}

func (w *Worker) agentCooperativeClose(logger log.Logger, job *data.Job,
	channel *data.Channel) error {
	logger = logger.Add("channel", channel)

	offering, err := w.offering(logger, channel.Offering)
	if err != nil {
		return err
	}

	agent, err := w.account(logger, channel.Agent)
	if err != nil {
		return err
	}

	offeringHash, err := w.toOfferingHashArr(logger, offering.Hash)
	if err != nil {
		return err
	}

	clientAddr, err := data.HexToAddress(channel.Client)
	if err != nil {
		logger.Error(err.Error())
		return ErrParseEthAddr
	}

	balance := big.NewInt(int64(channel.ReceiptBalance))
	block := uint32(channel.Block)

	closingHash := eth.BalanceClosingHash(clientAddr, w.pscAddr, block,
		offeringHash, balance)

	accKey, err := w.key(logger, agent.PrivateKey)
	if err != nil {
		return err
	}

	closingSig, err := crypto.Sign(closingHash, accKey)
	if err != nil {
		logger.Error(err.Error())
		return ErrSignClosingMsg
	}

	agentAddr, err := data.HexToAddress(channel.Agent)
	if err != nil {
		logger.Error(err.Error())
		return ErrParseEthAddr
	}

	if channel.ReceiptSignature == nil {
		return ErrNoReceiptSignature
	}

	balanceMsgSig, err := data.ToBytes(*channel.ReceiptSignature)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	auth := bind.NewKeyedTransactor(accKey)
	auth.GasLimit = w.gasConf.PSC.CooperativeClose

	tx, err := w.ethBack.CooperativeClose(auth, agentAddr,
		uint32(channel.Block), offeringHash, balance, balanceMsgSig,
		closingSig)
	if err != nil {
		logger.Add("GasLimit", auth.GasLimit,
			"GasPrice", auth.GasPrice).Error(err.Error())
		return ErrPSCCooperativeClose
	}

	return w.saveEthTX(logger, job, tx, "CooperativeClose", job.RelatedType,
		job.RelatedID, agent.EthAddr, data.HexFromBytes(w.pscAddr.Bytes()))
}

func (w *Worker) agentUpdateServiceStatus(logger log.Logger, job *data.Job,
	jobType string) (*data.Channel, error) {
	channel, err := w.relatedChannel(logger, job, jobType)
	if err != nil {
		return nil, err
	}

	switch jobType {
	case data.JobAgentPreServiceSuspend:
		channel.ServiceStatus = data.ServiceSuspended
	case data.JobAgentPreServiceTerminate:
		channel.ServiceStatus = data.ServiceTerminated
	case data.JobAgentPreServiceUnsuspend:
		channel.ServiceStatus = data.ServiceActive
	}

	changedTime := time.Now()
	channel.ServiceChangedTime = &changedTime

	if err = w.db.Update(channel); err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	return channel, nil
}

// AgentPreEndpointMsgCreate prepares endpoint message to be sent to client.
func (w *Worker) AgentPreEndpointMsgCreate(job *data.Job) error {
	logger := w.logger.Add("method", "AgentPreEndpointMsgCreate", "job", job)

	channel, err := w.relatedChannel(logger, job, data.JobAgentPreEndpointMsgCreate)
	if err != nil {
		return err
	}

	logger = logger.Add("channel", channel)

	msg, err := w.ept.EndpointMessage(channel.ID)
	if err != nil {
		logger.Error(err.Error())
		return ErrMakeEndpointMsg
	}

	logger = logger.Add("endpointMsg", msg)

	template, err := w.templateByHash(logger, msg.TemplateHash)
	if err != nil {
		return err
	}

	logger = logger.Add("template", template)

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	client, err := w.user(logger, channel.Client)
	if err != nil {
		return err
	}

	logger = logger.Add("client", client)

	clientPub, err := data.ToBytes(client.PublicKey)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	agent, err := w.account(logger, channel.Agent)
	if err != nil {
		return err
	}

	logger = logger.Add("agent", agent)

	agentKey, err := w.key(logger, agent.PrivateKey)
	if err != nil {
		return err
	}

	msgSealed, err := messages.AgentSeal(msgBytes, clientPub, agentKey)
	if err != nil {
		logger.Error(err.Error())
		return ErrEndpointMsgSeal
	}

	hash := crypto.Keccak256(msgSealed)

	var params []byte

	params, err = json.Marshal(msg.AdditionalParams)
	if err != nil {
		params = []byte("{}")
	}

	newEndpoint := &data.Endpoint{
		ID:                     util.NewUUID(),
		Template:               template.ID,
		Channel:                channel.ID,
		Hash:                   data.FromBytes(hash),
		RawMsg:                 data.FromBytes(msgSealed),
		Status:                 data.MsgUnpublished,
		ServiceEndpointAddress: pointer.ToString(msg.ServiceEndpointAddress),
		PaymentReceiverAddress: pointer.ToString(msg.PaymentReceiverAddress),
		Username:               pointer.ToString(msg.Username),
		Password:               pointer.ToString(msg.Password),
		AdditionalParams:       params,
	}

	tx, err := w.db.Begin()
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	if err = tx.Insert(newEndpoint); err != nil {
		logger.Error(err.Error())
		tx.Rollback()
		return err
	}

	salt, err := rand.Int(rand.Reader, big.NewInt(9*1e18))
	if err != nil {
		return err
	}

	passwordHash, err := data.HashPassword(msg.Password, string(salt.Uint64()))
	if err != nil {
		logger.Error(err.Error())
		return ErrGeneratePasswordHash
	}

	channel.Password = passwordHash
	channel.Salt = salt.Uint64()

	if err = tx.Update(channel); err != nil {
		logger.Error(err.Error())
		tx.Rollback()
		return ErrInternal
	}

	if err = tx.Commit(); err != nil {
		logger.Error(err.Error())
		tx.Rollback()
		return ErrInternal
	}

	return w.addJob(logger, data.JobAgentPreEndpointMsgSOMCPublish,
		data.JobEndpoint, newEndpoint.ID)
}

// AgentPreEndpointMsgSOMCPublish sends msg to somc and creates after job.
func (w *Worker) AgentPreEndpointMsgSOMCPublish(job *data.Job) error {
	logger := w.logger.Add("method", "AgentPreEndpointMsgSOMCPublish",
		"job", job)

	endpoint, err := w.relatedEndpoint(logger,
		job, data.JobAgentPreEndpointMsgSOMCPublish)
	if err != nil {
		return err
	}

	msg, err := data.ToBytes(endpoint.RawMsg)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	key, err := w.keyFromChannelData(logger, endpoint.Channel)
	if err != nil {
		return fmt.Errorf("failed to generate channel key: %v", err)
	}

	if err = w.somc.PublishEndpoint(key, msg); err != nil {
		logger.Error(err.Error())
		return ErrPublishEndpoint
	}

	endpoint.Status = data.MsgChPublished

	if err = w.db.Update(endpoint); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return w.addJob(logger, data.JobAgentAfterEndpointMsgSOMCPublish,
		data.JobChannel, endpoint.Channel)
}

// AgentAfterEndpointMsgSOMCPublish suspends service if some pre payment expected.
func (w *Worker) AgentAfterEndpointMsgSOMCPublish(job *data.Job) error {
	logger := w.logger.Add("method", "AgentAfterEndpointMsgSOMCPublish",
		"job", job)

	channel, err := w.relatedChannel(logger, job,
		data.JobAgentAfterEndpointMsgSOMCPublish)
	if err != nil {
		return err
	}

	offering, err := w.offering(logger, channel.Offering)
	if err != nil {
		return err
	}

	if offering.BillingType == data.BillingPrepaid ||
		offering.SetupPrice > 0 {
		channel.ServiceStatus = data.ServiceSuspended

	} else {
		channel.ServiceStatus = data.ServiceActive
	}

	changedTime := time.Now().Add(time.Minute)
	channel.ServiceChangedTime = &changedTime

	if err = w.db.Update(channel); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return nil
}

// AgentPreOfferingMsgBCPublish publishes offering to blockchain.
func (w *Worker) AgentPreOfferingMsgBCPublish(job *data.Job) error {
	logger := w.logger.Add("method", "AgentPreOfferingMsgBCPublish",
		"job", job)
	offering, err := w.relatedOffering(logger, job,
		data.JobAgentPreOfferingMsgBCPublish)
	if err != nil {
		return err
	}

	minDeposit := offering.MinUnits*offering.UnitPrice + offering.SetupPrice

	agent, err := w.account(logger, offering.Agent)
	if err != nil {
		return err
	}

	agentKey, err := w.key(logger, agent.PrivateKey)
	if err != nil {
		return err
	}

	template, err := w.template(logger, offering.Template)
	if err != nil {
		return err
	}

	msg := offer.OfferingMessage(agent, template, offering)

	logger = logger.Add("msg", msg)

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	packed, err := messages.PackWithSignature(msgBytes, agentKey)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	offering.RawMsg = data.FromBytes(packed)

	offeringHash := common.BytesToHash(crypto.Keccak256(packed))

	offering.Hash = data.FromBytes(offeringHash.Bytes())

	publishData, err := w.publishData(logger, job)
	if err != nil {
		return err
	}

	auth := bind.NewKeyedTransactor(agentKey)

	pscBalance, err := w.ethBack.PSCBalanceOf(&bind.CallOpts{}, auth.From)

	if err != nil {
		logger.Error(err.Error())
		return ErrPSCReturnBalance
	}

	totalDeposit := minDeposit * uint64(offering.Supply)
	if pscBalance.Uint64() < totalDeposit {
		return ErrInsufficientPSCBalance
	}

	ethAmount, err := w.ethBalance(logger, auth.From)
	if err != nil {
		return err
	}

	wantedEthBalance := auth.GasLimit * publishData.GasPrice
	if wantedEthBalance > ethAmount.Uint64() {
		return ErrInsufficientEthBalance
	}

	auth.GasLimit = w.gasConf.PSC.RegisterServiceOffering
	auth.GasPrice = big.NewInt(int64(publishData.GasPrice))

	tx, err := w.ethBack.RegisterServiceOffering(auth,
		[common.HashLength]byte(offeringHash),
		big.NewInt(int64(minDeposit)), offering.Supply)
	if err != nil {
		logger.Add("GasLimit", auth.GasLimit,
			"GasPrice", auth.GasPrice).Error(err.Error())
		return ErrPSCRegisterOffering
	}

	offering.Status = data.MsgBChainPublishing
	offering.OfferStatus = data.OfferRegister
	if err = w.db.Update(offering); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return w.saveEthTX(logger, job, tx, "RegisterServiceOffering", job.RelatedType,
		job.RelatedID, agent.EthAddr, data.HexFromBytes(w.pscAddr.Bytes()))
}

// AgentAfterOfferingMsgBCPublish updates offering status and creates
// somc publish job.
func (w *Worker) AgentAfterOfferingMsgBCPublish(job *data.Job) error {
	logger := w.logger.Add("method", "AgentAfterOfferingMsgBCPublish",
		"job", job)
	offering, err := w.relatedOffering(logger, job,
		data.JobAgentAfterOfferingMsgBCPublish)
	if err != nil {
		return err
	}

	offering.Status = data.MsgBChainPublished
	if err = w.db.Update(offering); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return w.addJob(logger, data.JobAgentPreOfferingMsgSOMCPublish,
		data.JobOffering, offering.ID)
}

// AgentPreOfferingMsgSOMCPublish publishes to somc and creates after job.
func (w *Worker) AgentPreOfferingMsgSOMCPublish(job *data.Job) error {
	logger := w.logger.Add("method", "AgentPreOfferingMsgSOMCPublish",
		"job", job)
	offering, err := w.relatedOffering(logger, job,
		data.JobAgentPreOfferingMsgSOMCPublish)
	if err != nil {
		return err
	}

	logger = logger.Add("offering", offering)

	packedMsgBytes, err := data.ToBytes(offering.RawMsg)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	if err = w.somc.PublishOffering(packedMsgBytes); err != nil {
		logger.Error(err.Error())
		return ErrPublishOffering
	}

	offering.Status = data.MsgChPublished
	if err = w.db.Update(offering); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	agent, err := w.account(logger, offering.Agent)
	if err != nil {
		return err
	}

	return w.addJob(logger, data.JobAccountUpdateBalances, data.JobAccount, agent.ID)
}

// AgentAfterOfferingDelete set offering status to `remove`
func (w *Worker) AgentAfterOfferingDelete(job *data.Job) error {
	return w.updateRelatedOffering(
		job, data.JobAgentAfterOfferingDelete, data.OfferRemove)
}

// AgentPreOfferingDelete calls psc remove an offering.
func (w *Worker) AgentPreOfferingDelete(job *data.Job) error {
	logger := w.logger.Add("method", "AgentPreOfferingDelete", "job", job)

	offering, err := w.relatedOffering(logger,
		job, data.JobAgentPreOfferingDelete)
	if err != nil {
		return err
	}

	if offering.OfferStatus != data.OfferRegister {
		return ErrOfferNotRegistered
	}

	jobDate, err := w.publishData(logger, job)
	if err != nil {
		return err
	}

	key, err := w.accountKey(logger, offering.Agent)
	if err != nil {
		return err
	}

	offeringHash, err := data.ToHash(offering.Hash)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	auth := bind.NewKeyedTransactor(key)
	auth.GasLimit = w.gasConf.PSC.RemoveServiceOffering
	auth.GasPrice = big.NewInt(int64(jobDate.GasPrice))

	tx, err := w.ethBack.PSCRemoveServiceOffering(auth, offeringHash)
	if err != nil {
		logger.Add("GasLimit", auth.GasLimit,
			"GasPrice", auth.GasPrice).Error(err.Error())
		return ErrPSCRemoveOffering
	}

	return w.saveEthTX(logger, job, tx, "RemoveServiceOffering",
		job.RelatedType, job.RelatedID, offering.Agent,
		data.HexFromBytes(w.pscAddr.Bytes()))
}

// AgentPreOfferingPopUp pop ups an offering.
func (w *Worker) AgentPreOfferingPopUp(job *data.Job) error {
	logger := w.logger.Add("method", "AgentPreOfferingPopUp", "job", job)
	offering, err := w.relatedOffering(logger,
		job, data.JobAgentPreOfferingPopUp)
	if err != nil {
		return err
	}

	if offering.OfferStatus != data.OfferRegister {
		return ErrOfferNotRegistered
	}

	jobDate, err := w.publishData(logger, job)
	if err != nil {
		return err
	}

	key, err := w.accountKey(logger, offering.Agent)
	if err != nil {
		return err
	}

	offeringHash, err := data.ToHash(offering.Hash)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	auth := bind.NewKeyedTransactor(key)
	auth.GasLimit = w.gasConf.PSC.PopupServiceOffering
	auth.GasPrice = big.NewInt(int64(jobDate.GasPrice))

	tx, err := w.ethBack.PSCPopupServiceOffering(auth, offeringHash)
	if err != nil {
		logger.Add("GasLimit", auth.GasLimit,
			"GasPrice", auth.GasPrice).Error(err.Error())
		return ErrPSCPopUpOffering
	}

	return w.saveEthTX(logger, job, tx, "PopupServiceOffering",
		job.RelatedType, job.RelatedID, offering.Agent,
		data.HexFromBytes(w.pscAddr.Bytes()))
}
