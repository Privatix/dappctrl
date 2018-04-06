package worker

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
)

// AgentAfterChannelCreate registers client and creates pre service create job.
func (w *Worker) AgentAfterChannelCreate(job *data.Job) error {
	channel, err := w.relatedChannel(job, data.JobAgentAfterChannelCreate)
	if err != nil {
		return err
	}

	tx, err := w.db.Begin()
	if err != nil {
		return err
	}

	channel.ChannelStatus = data.ChannelActive
	channel.ServiceStatus = data.ServicePending
	if err := tx.Update(channel); err != nil {
		tx.Rollback()
		return err
	}

	ethLogTx, err := w.ethLogTx(job)
	if err != nil {
		return err
	}

	client, err := w.newUser(ethLogTx)
	if err != nil {
		return err
	}

	if err := tx.Insert(client); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
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
		return err
	}

	agentAddr, err := data.ToAddress(channel.Agent)
	if err != nil {
		return err
	}

	clientAddr, err := data.ToAddress(channel.Client)
	if err != nil {
		return err
	}

	offering, err := w.offering(channel.Offering)
	if err != nil {
		return err
	}

	offeringHash, err := w.toHashArr(offering.Hash)
	if err != nil {
		return err
	}

	if agentAddr != logInput.agentAddr ||
		clientAddr != logInput.clientAddr ||
		offeringHash != logInput.offeringHash ||
		channel.Block != logInput.openBlockNum {
		return fmt.Errorf("channel mismatch")
	}

	channel.TotalDeposit += logInput.addedDeposit.Uint64()
	return w.db.Update(channel)
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
		return err
	}

	channel.ChannelStatus = data.ChannelInChallenge
	return w.db.Update(channel)
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
	return w.db.Update(channel)
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
		return err
	}

	balance := big.NewInt(int64(channel.ReceiptBalance))
	block := uint32(channel.Block)

	closingHash := eth.BalanceClosingHash(clientAddr, w.pscAddr, block,
		offeringHash, balance)

	accKey, err := w.key(agent.PrivateKey)
	if err != nil {
		return err
	}

	closingSig, err := crypto.Sign(closingHash, accKey)
	if err != nil {
		return err
	}

	agentAddr, err := data.ToAddress(channel.Agent)
	if err != nil {
		return err
	}

	balanceMsgSig, err := data.ToBytes(channel.ReceiptSignature)
	if err != nil {
		return err
	}

	auth := bind.NewKeyedTransactor(accKey)

	err = w.ethBack.CooperativeClose(auth, agentAddr, uint32(channel.Block),
		offeringHash, balance, balanceMsgSig,
		closingSig)
	if err != nil {
		return err
	}

	return w.addJob(data.JobAgentPreServiceTerminate, data.JobChannel,
		channel.ID)
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

	return w.db.Update(channel)
}

// AgentPreEndpointMessageCreate prepares endpoint message to be sent to client.
func (w *Worker) AgentPreEndpointMessageCreate(job *data.Job) error {
	// TODO
	return nil
}

// AgentPreEndpointMsgSOMCPublish sends msg to somc and creates after job.
func (w *Worker) AgentPreEndpointMsgSOMCPublish(job *data.Job) error {
	// TODO
	return nil
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

	if offering.BillingType == data.BillingPrepaid ||
		offering.SetupPrice > 0 {
		channel.ServiceStatus = data.ServiceSuspended
		return w.db.Update(channel)
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

	offeringHash, err := data.ToHash(offering.Hash)
	if err != nil {
		return err
	}

	agent, err := w.account(offering.Agent)
	if err != nil {
		return err
	}

	accKey, err := w.key(agent.PrivateKey)
	if err != nil {
		return err
	}

	auth := bind.NewKeyedTransactor(accKey)
	w.ethBack.RegisterServiceOffering(auth,
		[common.HashLength]byte(offeringHash),
		big.NewInt(int64(minDeposit)), offering.Supply)

	offering.Status = data.MsgBChainPublishing
	offering.OfferStatus = data.OfferRegister
	return w.db.Update(offering)
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
		return err
	}

	return w.addJob(data.JobAgentPreOfferingMsgSOMCPublish,
		data.JobOfferring, offering.ID)
}

// AgentPreOfferingMsgSOMCPublish publishes to somc and creates after job.
func (w *Worker) AgentPreOfferingMsgSOMCPublish(job *data.Job) error {
	// TODO
	offering, err := w.relatedOffering(job,
		data.JobAgentPreOfferingMsgSOMCPublish)
	if err != nil {
		return err
	}

	// agent, err := w.account(offering.Agent)
	// if err != nil {
	// 	return err
	// }

	// template, err := w.template(offering.Template)
	// if err != nil {
	// 	return err
	// }

	// msg := so.NewOfferingMessage(agent, template, offering)

	// key, err := w.key(agent.PrivateKey)
	// if err != nil {
	// 	return err
	// }

	// msgBytes, err := json.Marshal(msg)
	// if err != nil {
	// 	return err
	// }

	// packed, err := crypto.SignedMessage(key, msgBytes)
	// if err != nil {
	// 	return err
	// }

	// return w.somc.PublishOffering(packed)
	offering.Status = data.MsgChPublished
	if err = w.db.Update(offering); err != nil {
		return err
	}

	return w.addJob(data.JobAgentAfterOfferingMsgSOMCPublish,
		data.JobOfferring,
		offering.ID)
}

// AgentPreAccountAddBalanceApprove approve balance if amount exists.
func (w *Worker) AgentPreAccountAddBalanceApprove(job *data.Job) error {
	acc, err := w.relatedAccount(job,
		data.JobAgentPreAccountAddBalanceApprove)
	if err != nil {
		return err
	}

	jobData, err := w.balanceData(job)
	if err != nil {
		return err
	}

	addr, err := data.ToAddress(acc.EthAddr)
	if err != nil {
		return err
	}

	amount, err := w.ethBack.PTCBalanceOf(&bind.CallOpts{}, addr)
	if err != nil {
		return err
	}

	if amount.Uint64() < uint64(jobData.Amount) {
		return fmt.Errorf("not enough balance at ptc")
	}

	key, err := w.key(acc.PrivateKey)
	if err != nil {
		return err
	}

	auth := bind.NewKeyedTransactor(key)
	return w.ethBack.PTCApprove(auth,
		w.pscAddr, big.NewInt(int64(jobData.Amount)))
}

// AgentPreAccountAddBalance adds balance to psc.
func (w *Worker) AgentPreAccountAddBalance(job *data.Job) error {
	acc, err := w.relatedAccount(job, data.JobAgentPreAccountAddBalance)
	if err != nil {
		return err
	}

	jobData, err := w.balanceData(job)
	if err != nil {
		return err
	}

	key, err := w.key(acc.PrivateKey)
	if err != nil {
		return err
	}

	auth := bind.NewKeyedTransactor(key)
	return w.ethBack.PSCAddBalanceERC20(auth, big.NewInt(int64(jobData.Amount)))
}

// AgentAfterAccountAddBalance updates psc and ptc balance of an account.
func (w *Worker) AgentAfterAccountAddBalance(job *data.Job) error {
	return w.updateAccountBalances(job, data.JobAgentAfterAccountAddBalance)
}

// AgentPreAccountReturnBalance returns from psc to ptc.
func (w *Worker) AgentPreAccountReturnBalance(job *data.Job) error {
	acc, err := w.relatedAccount(job, data.JobAgentPreAccountReturnBalance)
	if err != nil {
		return err
	}

	key, err := w.key(acc.PrivateKey)
	if err != nil {
		return err
	}

	auth := bind.NewKeyedTransactor(key)

	amount, err := w.ethBack.PSCBalanceOf(&bind.CallOpts{}, auth.From)
	if err != nil {
		return err
	}

	jobData, err := w.balanceData(job)
	if err != nil {
		return err
	}

	if amount.Uint64() > uint64(jobData.Amount) {
		return fmt.Errorf("not enough psc balance")
	}

	return w.ethBack.PSCReturnBalanceERC20(auth, big.NewInt(int64(jobData.Amount)))
}

// AgentAfterAccountReturnBalance updates psc and ptc balance of an account.
func (w *Worker) AgentAfterAccountReturnBalance(job *data.Job) error {
	return w.updateAccountBalances(job, data.JobAgentAfterAccountReturnBalance)
}
