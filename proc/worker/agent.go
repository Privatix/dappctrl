package worker

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/messages"
	"github.com/privatix/dappctrl/messages/offer"
	"github.com/privatix/dappctrl/util"
)

// AgentAfterChannelCreate registers client and creates pre service create job.
func (w *Worker) AgentAfterChannelCreate(job *data.Job) error {
	err := w.validateJob(job, data.JobAgentAfterChannelCreate, data.JobChannel)
	if err != nil {
		return err
	}

	ethLog, err := w.ethLog(job)
	if err != nil {
		return err
	}

	ethLogTx, err := w.ethLogTx(ethLog)
	if err != nil {
		return err
	}

	client, err := w.newUser(ethLogTx)
	if err != nil {
		return fmt.Errorf("failed to make new client record: %v", err)
	}

	tx, err := w.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	if err := tx.Insert(client); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert %T: %v", client, err)
	}

	logChannelCreated, err := extractLogChannelCreated(ethLog)
	if err != nil {
		return fmt.Errorf("could not parse log: %v", err)
	}

	offering, err := w.offeringByHash(logChannelCreated.offeringHash)
	if err != nil {
		return fmt.Errorf("could not find offering by hash: %v", err)
	}

	channel := &data.Channel{
		ID:            job.RelatedID,
		Client:        data.FromBytes(logChannelCreated.clientAddr.Bytes()),
		Agent:         data.FromBytes(logChannelCreated.agentAddr.Bytes()),
		TotalDeposit:  logChannelCreated.deposit.Uint64(),
		ChannelStatus: data.ChannelActive,
		ServiceStatus: data.ServicePending,
		Offering:      offering.ID,
	}

	if err := tx.Insert(channel); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert %T: %v", channel, err)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to commit changes: %v", err)
	}

	return w.addJob(data.JobAgentPreEndpointMsgCreate,
		data.JobChannel, channel.ID)
}

// AgentAfterChannelTopUp updates deposit of a channel.
func (w *Worker) AgentAfterChannelTopUp(job *data.Job) error {
	channel, err := w.relatedChannel(job, data.JobAgentAfterChannelTopUp)
	if err != nil {
		return err
	}

	ethLog, err := w.ethLog(job)
	if err != nil {
		return err
	}

	logInput, err := extractLogChannelToppedUp(ethLog)
	if err != nil {
		return fmt.Errorf("could not parse log: %v", err)
	}

	agentAddr, err := data.ToAddress(channel.Agent)
	if err != nil {
		return fmt.Errorf("failed to parse agent addr: %v", err)
	}

	clientAddr, err := data.ToAddress(channel.Client)
	if err != nil {
		return fmt.Errorf("failed to parse client addr: %v", err)
	}

	offering, err := w.offering(channel.Offering)
	if err != nil {
		return err
	}

	offeringHash, err := w.toHashArr(offering.Hash)
	if err != nil {
		return fmt.Errorf("could not parse offering hash: %v", err)
	}

	if agentAddr != logInput.agentAddr ||
		clientAddr != logInput.clientAddr ||
		offeringHash != logInput.offeringHash ||
		channel.Block != logInput.openBlockNum {
		return fmt.Errorf("related channel does not correspond to log input")
	}

	channel.TotalDeposit += logInput.addedDeposit.Uint64()
	if err = w.db.Update(channel); err != nil {
		return fmt.Errorf("could not update channels deposit: %v", err)
	}

	return nil
}

// AgentAfterUncooperativeCloseRequest sets channel's status to challenge period.
func (w *Worker) AgentAfterUncooperativeCloseRequest(job *data.Job) error {
	channel, err := w.relatedChannel(job,
		data.JobAgentAfterUncooperativeCloseRequest)
	if err != nil {
		return err
	}

	var jobType string

	if channel.ReceiptBalance > 0 {
		jobType = data.JobAgentPreCooperativeClose
	} else {
		jobType = data.JobAgentPreServiceTerminate
	}

	if err = w.addJob(jobType, data.JobChannel, channel.ID); err != nil {
		return fmt.Errorf("could not add %s job: %v", jobType, err)
	}

	channel.ChannelStatus = data.ChannelInChallenge
	if err = w.db.Update(channel); err != nil {
		return fmt.Errorf("could not update channel's status: %v", err)
	}

	return nil
}

// AgentAfterUncooperativeClose marks channel closed uncoop.
func (w *Worker) AgentAfterUncooperativeClose(job *data.Job) error {
	channel, err := w.relatedChannel(job,
		data.JobAgentAfterUncooperativeClose)
	if err != nil {
		return err
	}

	if err = w.addJob(data.JobAgentPreServiceTerminate, data.JobChannel,
		channel.ID); err != nil {
		return err
	}

	channel.ChannelStatus = data.ChannelClosedUncoop
	if err = w.db.Update(channel); err != nil {
		return fmt.Errorf("could not update channel's status: %v", err)
	}

	return nil
}

// AgentPreCooperativeClose call contract cooperative close method and trigger
// service terminate job.
func (w *Worker) AgentPreCooperativeClose(job *data.Job) error {
	channel, err := w.relatedChannel(job, data.JobAgentPreCooperativeClose)
	if err != nil {
		return err
	}

	offering, err := w.offering(channel.Offering)
	if err != nil {
		return err
	}

	agent, err := w.account(channel.Agent)
	if err != nil {
		return err
	}

	offeringHash, err := w.toHashArr(offering.Hash)
	if err != nil {
		return err
	}

	clientAddr, err := data.ToAddress(channel.Client)
	if err != nil {
		return fmt.Errorf("unable to parse client addr: %v", err)
	}

	balance := big.NewInt(int64(channel.ReceiptBalance))
	block := uint32(channel.Block)

	closingHash := eth.BalanceClosingHash(clientAddr, w.pscAddr, block,
		offeringHash, balance)

	accKey, err := w.key(agent.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to parse agent's key: %v", err)
	}

	closingSig, err := crypto.Sign(closingHash, accKey)
	if err != nil {
		return fmt.Errorf("could not sign closing msg: %v", err)
	}

	agentAddr, err := data.ToAddress(channel.Agent)
	if err != nil {
		return fmt.Errorf("unable to parse agent's address: %v", err)
	}

	if channel.ReceiptSignature == nil {
		return fmt.Errorf("no receipt signature in channel")
	}

	balanceMsgSig, err := data.ToBytes(*channel.ReceiptSignature)
	if err != nil {
		return fmt.Errorf("unable to decode receipt signature: %v", err)
	}

	auth := bind.NewKeyedTransactor(accKey)
	auth.GasLimit = w.gasConf.PSC.CooperativeClose

	tx, err := w.ethBack.CooperativeClose(auth, agentAddr,
		uint32(channel.Block), offeringHash, balance, balanceMsgSig,
		closingSig)
	if err != nil {
		return fmt.Errorf("could not cooperative close: %v", err)
	}

	if err = w.addJob(data.JobAgentPreServiceTerminate, data.JobChannel,
		channel.ID); err != nil {
		return fmt.Errorf("could not add job: %v", err)
	}

	return w.saveEthTX(job, tx, "CooperativeClose", job.RelatedType,
		job.RelatedID, agent.EthAddr, data.FromBytes(w.pscAddr.Bytes()))
}

// AgentAfterCooperativeClose marks channel as closed coop.
func (w *Worker) AgentAfterCooperativeClose(job *data.Job) error {
	channel, err := w.relatedChannel(job, data.JobAgentAfterCooperativeClose)
	if err != nil {
		return err
	}

	channel.ChannelStatus = data.ChannelClosedCoop
	return w.db.Update(channel)
}

// AgentPreServiceSuspend marks service as suspended.
func (w *Worker) AgentPreServiceSuspend(job *data.Job) error {
	return w.agentUpdateServiceStatus(job, data.JobAgentPreServiceSuspend)
}

// AgentPreServiceUnsuspend marks service as active.
func (w *Worker) AgentPreServiceUnsuspend(job *data.Job) error {
	return w.agentUpdateServiceStatus(job, data.JobAgentPreServiceUnsuspend)
}

// AgentPreServiceTerminate marks service as active.
func (w *Worker) AgentPreServiceTerminate(job *data.Job) error {
	return w.agentUpdateServiceStatus(job, data.JobAgentPreServiceTerminate)
}

func (w *Worker) agentUpdateServiceStatus(job *data.Job, jobType string) error {
	channel, err := w.relatedChannel(job, jobType)
	if err != nil {
		return err
	}

	switch jobType {
	case data.JobAgentPreServiceSuspend:
		channel.ServiceStatus = data.ServiceSuspended
	case data.JobAgentPreServiceTerminate:
		channel.ServiceStatus = data.ServiceTerminated
	case data.JobAgentPreServiceUnsuspend:
		channel.ServiceStatus = data.ServiceActive
	}

	if err = w.db.Update(channel); err != nil {
		return fmt.Errorf("could not update service status: %v", err)
	}

	return nil
}

// AgentPreEndpointMsgCreate prepares endpoint message to be sent to client.
func (w *Worker) AgentPreEndpointMsgCreate(job *data.Job) error {
	channel, err := w.relatedChannel(job, data.JobAgentPreEndpointMsgCreate)
	if err != nil {
		return err
	}

	// TODO: move timeout to conf.
	msg, err := w.ept.EndpointMessage(channel.ID, time.Second)
	if err != nil {
		return fmt.Errorf("could not make endpoint message: %v", err)
	}

	template, err := w.templateByHash(msg.TemplateHash)
	if err != nil {
		return err
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("could not marshal endpoint msg: %v", err)
	}

	client, err := w.user(channel.Client)
	if err != nil {
		return fmt.Errorf("could not find channel's client: %v", err)
	}

	clientPub, err := data.ToBytes(client.PublicKey)
	if err != nil {
		return fmt.Errorf("unable to parse client's pub key: %v", err)
	}

	agent, err := w.account(channel.Agent)
	if err != nil {
		return fmt.Errorf("could not find channel's agent: %v", err)
	}

	agentKey, err := w.key(agent.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to parse agent's priv key: %v", err)
	}

	msgSealed, err := messages.AgentSeal(msgBytes, clientPub, agentKey)
	if err != nil {
		return fmt.Errorf("could not seal endpoint message: %v", err)
	}

	hash := crypto.Keccak256(msgSealed)

	newEndpoint := &data.Endpoint{
		ID:               util.NewUUID(),
		Template:         template.ID,
		Channel:          channel.ID,
		Hash:             data.FromBytes(hash),
		RawMsg:           data.FromBytes(msgSealed),
		Status:           data.MsgUnpublished,
		AdditionalParams: []byte("{}"),
	}

	tx, err := w.db.Begin()
	if err != nil {
		return fmt.Errorf("could not start db transaction: %v", err)
	}

	if err = tx.Insert(newEndpoint); err != nil {
		tx.Rollback()
		return fmt.Errorf("could not insert %T: %v", newEndpoint, err)
	}

	salt, err := rand.Int(rand.Reader, big.NewInt(9*1e18))
	if err != nil {
		return fmt.Errorf("failed to generate salt: %v", err)
	}

	passwordHash, err := data.HashPassword(msg.Password, string(salt.Uint64()))
	if err != nil {
		return fmt.Errorf("failed to generate password hash: %v", err)
	}

	channel.Password = passwordHash
	channel.Salt = salt.Uint64()

	if err = tx.Update(channel); err != nil {
		tx.Rollback()
		return fmt.Errorf("could not update %T: %v", channel, err)
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit changes: %v", err)
	}

	return w.addJob(data.JobAgentPreEndpointMsgSOMCPublish,
		data.JobEndpoint, newEndpoint.ID)
}

// AgentPreEndpointMsgSOMCPublish sends msg to somc and creates after job.
func (w *Worker) AgentPreEndpointMsgSOMCPublish(job *data.Job) error {
	endpoint, err := w.relatedEndpoint(job, data.JobAgentPreEndpointMsgSOMCPublish)
	if err != nil {
		return err
	}

	msg, err := data.ToBytes(endpoint.RawMsg)
	if err != nil {
		return fmt.Errorf("unable to parse endpoint's raw msg: %v", err)
	}

	if err = w.somc.PublishEndpoint(endpoint.Channel, msg); err != nil {
		return fmt.Errorf("could not publish endpoint msg: %v", err)
	}

	endpoint.Status = data.MsgChPublished

	if err = w.db.Update(endpoint); err != nil {
		return fmt.Errorf("could not update %T: %v", endpoint, err)
	}

	return w.addJob(data.JobAgentAfterEndpointMsgSOMCPublish,
		data.JobChannel, endpoint.Channel)
}

// AgentAfterEndpointMsgSOMCPublish suspends service if some pre payment expected.
func (w *Worker) AgentAfterEndpointMsgSOMCPublish(job *data.Job) error {
	channel, err := w.relatedChannel(job,
		data.JobAgentAfterEndpointMsgSOMCPublish)
	if err != nil {
		return err
	}

	offering, err := w.offering(channel.Offering)
	if err != nil {
		return err
	}

	if offering.BillingType == data.BillingPrepaid || offering.SetupPrice > 0 {
		channel.ServiceStatus = data.ServiceSuspended
		if err = w.db.Update(channel); err != nil {
			return fmt.Errorf("failed to update %T: %v", channel, err)
		}
	}

	return nil
}

// AgentPreOfferingMsgBCPublish publishes offering to blockchain.
func (w *Worker) AgentPreOfferingMsgBCPublish(job *data.Job) error {
	offering, err := w.relatedOffering(job,
		data.JobAgentPreOfferingMsgBCPublish)
	if err != nil {
		return err
	}

	minDeposit := offering.MinUnits*offering.UnitPrice + offering.SetupPrice

	agent, err := w.account(offering.Agent)
	if err != nil {
		return fmt.Errorf("could not find offering's agent: %v", err)
	}

	agentKey, err := w.key(agent.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to parse agent's priv key: %v", err)
	}

	template, err := w.template(offering.Template)
	if err != nil {
		return fmt.Errorf("could not find offering's template: %v", err)
	}

	msg := offer.OfferingMessage(agent, template, offering)

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal offering msg: %v", err)
	}

	packed, err := messages.PackWithSignature(msgBytes, agentKey)
	if err != nil {
		return fmt.Errorf("failed to pack msg with signature: %v", err)
	}

	offering.RawMsg = data.FromBytes(packed)

	offeringHash := common.BytesToHash(crypto.Keccak256(packed))

	offering.Hash = data.FromBytes(offeringHash.Bytes())

	publishData, err := w.publishData(job)
	if err != nil {
		return err
	}

	auth := bind.NewKeyedTransactor(agentKey)

	pscBalance, err := w.ethBack.PSCBalanceOf(&bind.CallOpts{}, auth.From)

	if err != nil {
		return fmt.Errorf("failed to get psc balance: %v", err)
	}

	totalDeposit := minDeposit * uint64(offering.Supply)
	if pscBalance.Uint64() < totalDeposit {
		return fmt.Errorf("failed to publish: insufficient psc balance")
	}

	ethAmount, err := w.ethBalance(auth.From)
	if err != nil {
		return fmt.Errorf("failed to publish: %v", err)
	}

	wantedEthBalance := auth.GasLimit * publishData.GasPrice
	if wantedEthBalance > ethAmount.Uint64() {
		return fmt.Errorf("failed to publish: insufficient"+
			"eth balance, wanted %v, got: %v", wantedEthBalance,
			ethAmount.Uint64())
	}

	auth.GasPrice = big.NewInt(int64(publishData.GasPrice))
	auth.GasLimit = w.gasConf.PSC.RegisterServiceOffering

	tx, err := w.ethBack.RegisterServiceOffering(auth,
		[common.HashLength]byte(offeringHash),
		big.NewInt(int64(minDeposit)), offering.Supply)
	if err != nil {
		return err
	}

	offering.Status = data.MsgBChainPublishing
	offering.OfferStatus = data.OfferRegister
	if err = w.db.Update(offering); err != nil {
		return fmt.Errorf("failed to update offering: %v", err)
	}

	return w.saveEthTX(job, tx, "RegisterServiceOffering", job.RelatedType,
		job.RelatedID, agent.EthAddr, data.FromBytes(w.pscAddr.Bytes()))
}

// AgentAfterOfferingMsgBCPublish updates offering status and creates
// somc publish job.
func (w *Worker) AgentAfterOfferingMsgBCPublish(job *data.Job) error {
	offering, err := w.relatedOffering(job,
		data.JobAgentAfterOfferingMsgBCPublish)
	if err != nil {
		return err
	}

	offering.Status = data.MsgBChainPublished
	if err = w.db.Update(offering); err != nil {
		return fmt.Errorf("could not update %T: %v", offering, err)
	}

	return w.addJob(data.JobAgentPreOfferingMsgSOMCPublish,
		data.JobOfferring, offering.ID)
}

// AgentPreOfferingMsgSOMCPublish publishes to somc and creates after job.
func (w *Worker) AgentPreOfferingMsgSOMCPublish(job *data.Job) error {
	offering, err := w.relatedOffering(job,
		data.JobAgentPreOfferingMsgSOMCPublish)
	if err != nil {
		return err
	}

	offering.Status = data.MsgChPublished

	if err = w.db.Update(offering); err != nil {
		return fmt.Errorf("could not update %T: %v", offering, err)
	}

	packedMsgBytes, err := data.ToBytes(offering.RawMsg)
	if err != nil {
		return fmt.Errorf("failed to decode offering's raw msg: %v", err)
	}

	if err = w.somc.PublishOffering(packedMsgBytes); err != nil {
		return fmt.Errorf("could not publish offering: %v", err)
	}

	if err = w.db.Update(offering); err != nil {
		return fmt.Errorf("could not update %T: %v", offering, err)
	}

	return nil
}
