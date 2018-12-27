package worker

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"math/big"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/country"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/messages"
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
		Block:         uint32(ethLog.Block),
	}

	if err := tx.Insert(channel); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	if err := w.addJob(logger, tx, data.JobAgentPreEndpointMsgCreate,
		data.JobChannel, channel.ID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		logger.Error("unable to commit changes: " + err.Error())
		return ErrInternal
	}

	return nil
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

func (w *Worker) incrementCurrentSupply(logger log.Logger,
	db *reform.Querier, pk string) error {
	offering, err := w.offering(logger, pk)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	offering.CurrentSupply++
	if err := db.Update(offering); err != nil {
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

	if err := w.incrementCurrentSupply(logger,
		w.db.Querier, channel.Offering); err != nil {
		return err
	}

	agent, err := w.account(logger, channel.Agent)
	if err != nil {
		return err
	}

	return w.addJob(logger, nil,
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

	return w.db.InTransaction(func(tx *reform.TX) error {
		channel.ChannelStatus = data.ChannelClosedCoop
		if err := tx.Update(channel); err != nil {
			logger.Error(err.Error())
			return ErrInternal
		}

		if err := w.incrementCurrentSupply(logger, tx.Querier,
			channel.Offering); err != nil {
			return err
		}

		agent, err := w.account(logger, channel.Agent)
		if err != nil {
			return err
		}

		return w.addJob(logger, tx,
			data.JobAccountUpdateBalances, data.JobAccount, agent.ID)
	})
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

	if channel.ReceiptBalance == 0 {
		return nil
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

	balance := new(big.Int).SetUint64(channel.ReceiptBalance)
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
		Hash:                   data.HexFromBytes(hash),
		RawMsg:                 data.FromBytes(msgSealed),
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

	offering, err := w.offering(logger, channel.Offering)
	if err != nil {
		return err
	}

	channel.Password = passwordHash
	channel.Salt = salt.Uint64()

	if offering.BillingType == data.BillingPrepaid ||
		offering.SetupPrice > 0 {
		channel.ServiceStatus = data.ServiceSuspended

	} else {
		channel.ServiceStatus = data.ServiceActive
	}
	changedTime := time.Now().Add(time.Minute)
	channel.ServiceChangedTime = &changedTime

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

	minDeposit := data.MinDeposit(offering)

	agent, err := w.account(logger, offering.Agent)
	if err != nil {
		return err
	}

	agentKey, err := w.key(logger, agent.PrivateKey)
	if err != nil {
		return err
	}

	product, err := w.productByPK(logger, offering.Product)
	if err != nil {
		return err
	}

	offering.Country = country.UndefinedCountry

	if product.Country != nil && len(*product.Country) == 2 {
		offering.Country = *product.Country
	}

	publishData, err := w.publishData(logger, job)
	if err != nil {
		return err
	}

	offeringHash, err := data.HexToBytes(offering.Hash)
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
	auth.GasPrice = new(big.Int).SetUint64(publishData.GasPrice)

	if w.torHostName == "" {
		return ErrTorNoSet
	}

	offering.SOMCType = data.OfferingSOMCTor
	offering.SOMCData = w.torHostName

	tx, err := w.ethBack.RegisterServiceOffering(auth,
		[common.HashLength]byte(common.BytesToHash(offeringHash)),
		new(big.Int).SetUint64(minDeposit), offering.Supply,
		offering.SOMCType, offering.SOMCData)
	if err != nil {
		logger.Add("GasLimit", auth.GasLimit,
			"GasPrice", auth.GasPrice).Error(err.Error())
		return ErrPSCRegisterOffering
	}

	offering.Status = data.MsgBChainPublishing
	offering.OfferStatus = data.OfferRegistering
	if err = w.db.Update(offering); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return w.saveEthTX(logger, job, tx, "RegisterServiceOffering",
		job.RelatedType, job.RelatedID, agent.EthAddr,
		data.HexFromBytes(w.pscAddr.Bytes()))
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

	ethLog, err := w.ethLog(logger, job)
	if err != nil {
		return err
	}

	logger = logger.Add("ethLog", ethLog)

	offering.Status = data.MsgBChainPublished
	offering.OfferStatus = data.OfferRegistered
	offering.BlockNumberUpdated = ethLog.Block
	if err = w.db.Update(offering); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	agent, err := w.account(logger, offering.Agent)
	if err != nil {
		return err
	}

	return w.addJob(logger, nil,
		data.JobAccountUpdateBalances, data.JobAccount, agent.ID)
}

// AgentAfterOfferingDelete set offering status to `remove`
func (w *Worker) AgentAfterOfferingDelete(job *data.Job) error {
	logger := w.logger.Add(
		"method", "AgentAfterOfferingDelete", "job", job)

	offering, err := w.relatedOffering(
		logger, job, data.JobAgentAfterOfferingDelete)
	if err != nil {
		return err
	}
	offering.OfferStatus = data.OfferRemoved

	if err := w.saveRecord(logger, w.db.Querier, offering); err != nil {
		return err
	}

	agent, err := w.account(logger, offering.Agent)
	if err != nil {
		return err
	}

	return w.addJob(logger, nil,
		data.JobAccountUpdateBalances, data.JobAccount, agent.ID)
}

// AgentPreOfferingDelete calls psc remove an offering.
func (w *Worker) AgentPreOfferingDelete(job *data.Job) error {
	logger := w.logger.Add("method", "AgentPreOfferingDelete", "job", job)

	offering, err := w.relatedOffering(logger,
		job, data.JobAgentPreOfferingDelete)
	if err != nil {
		return err
	}

	if offering.OfferStatus != data.OfferRegistered &&
		offering.OfferStatus != data.OfferPoppedUp {
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

	offeringHash, err := data.HexToHash(offering.Hash)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	err = w.checkInPeriod(logger, offeringHash, data.SettingsPeriodRemove,
		ErrOfferingDeletePeriodIsNotOver)
	if err != nil {
		return err
	}

	auth := bind.NewKeyedTransactor(key)
	auth.GasLimit = w.gasConf.PSC.RemoveServiceOffering
	auth.GasPrice = new(big.Int).SetUint64(jobDate.GasPrice)

	tx, err := w.ethBack.PSCRemoveServiceOffering(auth, offeringHash)
	if err != nil {
		logger.Add("GasLimit", auth.GasLimit,
			"GasPrice", auth.GasPrice).Error(err.Error())
		return ErrPSCRemoveOffering
	}

	offering.OfferStatus = data.OfferRemoving
	if err := w.saveRecord(logger, w.db.Querier,
		offering); err != nil {
		return err
	}

	return w.saveEthTX(logger, job, tx, "RemoveServiceOffering",
		job.RelatedType, job.RelatedID, offering.Agent,
		data.HexFromBytes(w.pscAddr.Bytes()))
}

func (w *Worker) agentOfferingPopUpFindRelatedJobs(
	logger log.Logger, id, jobID string) error {
	query := `SELECT count(*)
                    FROM jobs
                   WHERE (jobs.type = $1
                         OR jobs.type = $2)
                         AND jobs.status = $3
			 AND jobs.related_id = $4
			 AND jobs.id != $5;`

	var count uint64
	err := w.db.QueryRow(query, data.JobAgentPreOfferingDelete,
		data.JobAgentPreOfferingPopUp, data.JobActive, id,
		jobID).Scan(&count)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	if count != 0 {
		return ErrUncompletedJobsExists
	}

	return err
}

// checkInPeriod checks an offering being in period specified by periodKey.
func (w *Worker) checkInPeriod(logger log.Logger, hash common.Hash,
	periodKey string, periodErr error) error {
	updateBlockNumber, err := w.getOfferingBlockNumber(logger, hash)
	if err != nil {
		return err
	}

	lastBlock, err := w.ethBack.LatestBlockNumber(context.Background())
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	removePeriod, err := data.ReadUintSetting(w.db.Querier, periodKey)
	if err != nil {
		return periodErr
	}

	if uint64(updateBlockNumber)+uint64(removePeriod) > lastBlock.Uint64() {
		return periodErr
	}

	return nil
}

func (w *Worker) getOfferingBlockNumber(logger log.Logger,
	hash common.Hash) (uint32, error) {

	_, _, _, _, updateBlockNumber, active, err := w.ethBack.PSCGetOfferingInfo(
		&bind.CallOpts{}, hash)
	if err != nil {
		logger.Error(err.Error())
		return 0, ErrInternal
	}

	if !active {
		return 0, ErrOfferingNotActive
	}

	return updateBlockNumber, err
}

// AgentPreOfferingPopUp pop ups an offering.
func (w *Worker) AgentPreOfferingPopUp(job *data.Job) error {
	logger := w.logger.Add("method", "AgentPreOfferingPopUp", "job", job)
	offering, err := w.relatedOffering(logger,
		job, data.JobAgentPreOfferingPopUp)
	if err != nil {
		return err
	}

	logger = logger.Add("offering", offering.ID)

	if offering.OfferStatus != data.OfferRegistered &&
		offering.OfferStatus != data.OfferPoppedUp {
		return ErrOfferNotRegistered
	}

	jobDate, err := w.publishData(logger, job)
	if err != nil {
		return err
	}

	offeringHash, err := data.HexToHash(offering.Hash)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	err = w.agentOfferingPopUpFindRelatedJobs(logger, offering.ID, job.ID)
	if err != nil {
		return err
	}

	err = w.checkInPeriod(logger, offeringHash, data.SettingsPeriodPopUp,
		ErrPopUpPeriodIsNotOver)
	if err != nil {
		return err
	}

	key, err := w.accountKey(logger, offering.Agent)
	if err != nil {
		return err
	}

	auth := bind.NewKeyedTransactor(key)
	auth.GasLimit = w.gasConf.PSC.PopupServiceOffering
	auth.GasPrice = new(big.Int).SetUint64(jobDate.GasPrice)

	tx, err := w.ethBack.PSCPopupServiceOffering(auth, offeringHash,
		offering.SOMCType, offering.SOMCData)
	if err != nil {
		logger.Add("GasLimit", auth.GasLimit,
			"GasPrice", auth.GasPrice).Error(err.Error())
		return ErrPSCPopUpOffering
	}

	offering.OfferStatus = data.OfferPoppingUp
	if err := w.saveRecord(logger, w.db.Querier, offering); err != nil {
		return err
	}

	return w.saveEthTX(logger, job, tx, "PopupServiceOffering",
		job.RelatedType, job.RelatedID, offering.Agent,
		data.HexFromBytes(w.pscAddr.Bytes()))
}

// AgentAfterOfferingPopUp updates the block number
// when the offering was updated.
func (w *Worker) AgentAfterOfferingPopUp(job *data.Job) error {
	logger := w.logger.Add("method", "AgentAfterOfferingPopUp", "job", job)

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
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	offering.BlockNumberUpdated = ethLog.Block
	offering.OfferStatus = data.OfferPoppedUp

	return w.saveRecord(logger, w.db.Querier, &offering)
}
