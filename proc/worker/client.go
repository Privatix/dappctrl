package worker

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/privatix/dappctrl/data"
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
	if err := data.Insert(w.db, &ch); err != nil {
		return err
	}

	agentAddr, err := data.ToAddress(offer.Agent)
	if err != nil {
		return err
	}

	tx, err := w.ethBack.PSCCreateChannel(&bind.TransactOpts{},
		agentAddr, offerHash, big.NewInt(int64(deposit)))
	if err != nil {
		return err
	}

	return w.saveEthTX(job, tx, "CreateChannel",
		data.JobChannel, ch.ID, acc.EthAddr, offer.Agent)
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
	err := data.FindByPrimaryKeyTo(w.db, &acc, jdata.Account)
	if err != nil {
		return err
	}

	var offer data.Offering
	err = data.FindByPrimaryKeyTo(w.db, &offer, jdata.Oferring)
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

func (w *Worker) ClientAfterChannelCreate(job *data.Job) error {
	var ch data.Channel
	err := data.FindByPrimaryKeyTo(w.db, &ch, job.RelatedID)
	if err != nil {
		return err
	}

	ch.ChannelStatus = data.ChannelActive
	if err = data.Save(w.db, &ch); err != nil {
		return err
	}

	return w.addJob(data.JobClientPreEndpointMsgSOMCGet,
		data.JobChannel, ch.ID)
}
