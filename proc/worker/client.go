package worker

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/privatix/dappctrl/data"
)

type clientPreChannelCreateData struct {
	Account  string `json:"account"`
	Oferring string `json:"offering"`
}

// ClientPreChannelCreate creates channel.
func (w *Worker) ClientPreChannelCreate(job *data.Job) error {
	var jdata clientPreChannelCreateData
	if err := parseJobData(job, &jdata); err != nil {
		return err
	}

	var acc data.Account
	err := data.FindByPrimaryKeyTo(w.db, &acc, jdata.Account)
	if err != nil {
		return err
	}

	addr, err := data.ToAddress(acc.EthAddr)
	if err != nil {
		return err
	}

	amount, err := w.ethBack.PSCBalanceOf(&bind.CallOpts{}, addr)
	if err != nil {
		return err
	}

	var offer data.Offering
	err = data.FindByPrimaryKeyTo(w.db, &offer, jdata.Oferring)
	if err != nil {
		return err
	}

	deposit := offer.UnitPrice*offer.MinUnits + offer.SetupPrice
	if amount.Uint64() < deposit {
		msg := "not enough PSC balance (%d) for offering %s"
		return fmt.Errorf(msg, amount.Uint64(), offer.ID)
	}

	hash, err := data.ToHash(offer.Hash)
	if err != nil {
		return err
	}

	supply, err := w.ethBack.PSCOfferingSupply(&bind.CallOpts{}, hash)
	if err != nil {
		return err
	}

	if supply == 0 {
		return fmt.Errorf("no supply for offering %s", offer.ID)
	}

	ch := data.Channel{
		ID:            job.RelatedID,
		Block:         0,
		ChannelStatus: data.ChannelPending,
		ServiceStatus: data.ServicePending,
	}
	if err := data.Insert(w.db, &ch); err != nil {
		return err
	}

	agentAddr, err := data.ToAddress(offer.Agent)
	if err != nil {
		return err
	}

	tx, err := w.ethBack.PSCCreateChannel(&bind.TransactOpts{},
		agentAddr, hash, big.NewInt(int64(deposit)))
	if err != nil {
		return err
	}

	return w.saveEthTX(job, tx, "CreateChannel",
		data.JobChannel, ch.ID, acc.EthAddr, offer.Agent)
}
