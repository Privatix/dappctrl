package worker

import (
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/privatix/dappctrl/data"
)

// PreAccountAddBalanceApprove approve balance if amount exists.
func (w *Worker) PreAccountAddBalanceApprove(job *data.Job) error {
	acc, err := w.relatedAccount(job,
		data.JobPreAccountAddBalanceApprove)
	if err != nil {
		return err
	}

	jobData, err := w.balanceData(job)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %v", err)
	}

	addr, err := data.ToAddress(acc.EthAddr)
	if err != nil {
		return fmt.Errorf("unable to parse account's addr: %v", err)
	}

	amount, err := w.ethBack.PTCBalanceOf(&bind.CallOpts{}, addr)
	if err != nil {
		return fmt.Errorf("could not get account's ptc balance: %v", err)
	}

	if amount.Uint64() < uint64(jobData.Amount) {
		return fmt.Errorf("not enough balance at ptc")
	}

	key, err := w.key(acc.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to parse account's priv key: %v", err)
	}

	auth := bind.NewKeyedTransactor(key)
	err = w.ethBack.PTCIncreaseApproval(auth,
		w.pscAddr, big.NewInt(int64(jobData.Amount)))
	if err != nil {
		return fmt.Errorf("could not ptc increase approve: %v", err)
	}

	return nil
}

// PreAccountAddBalance adds balance to psc.
func (w *Worker) PreAccountAddBalance(job *data.Job) error {
	acc, err := w.relatedAccount(job, data.JobPreAccountAddBalance)
	if err != nil {
		return err
	}

	jobData, err := w.balanceData(job)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %v", err)
	}

	key, err := w.key(acc.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to parse account's priv key: %v", err)
	}

	auth := bind.NewKeyedTransactor(key)
	err = w.ethBack.PSCAddBalanceERC20(auth, big.NewInt(int64(jobData.Amount)))
	if err != nil {
		return fmt.Errorf("could not add balance to psc: %v", err)
	}

	return nil
}

// AfterAccountAddBalance updates psc and ptc balance of an account.
func (w *Worker) AfterAccountAddBalance(job *data.Job) error {
	acc, err := w.relatedAccount(job, data.JobAfterAccountAddBalance)
	if err != nil {
		return err
	}

	return w.updateAccountBalances(acc)
}

// PreAccountReturnBalance returns from psc to ptc.
func (w *Worker) PreAccountReturnBalance(job *data.Job) error {
	acc, err := w.relatedAccount(job, data.JobPreAccountReturnBalance)
	if err != nil {
		return err
	}

	key, err := w.key(acc.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to parse account's priv key: %v", err)
	}

	auth := bind.NewKeyedTransactor(key)

	amount, err := w.ethBack.PSCBalanceOf(&bind.CallOpts{}, auth.From)
	if err != nil {
		return fmt.Errorf("could not get account's psc balance: %v", err)
	}

	jobData, err := w.balanceData(job)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %v", err)
	}

	if amount.Uint64() > uint64(jobData.Amount) {
		return fmt.Errorf("not enough psc balance")
	}

	err = w.ethBack.PSCReturnBalanceERC20(auth,
		big.NewInt(int64(jobData.Amount)))
	if err != nil {
		return fmt.Errorf("could not return balance from psc: %v", err)
	}

	return nil
}

// AfterAccountReturnBalance updates psc and ptc balance of an account.
func (w *Worker) AfterAccountReturnBalance(job *data.Job) error {
	acc, err := w.relatedAccount(job, data.JobAfterAccountReturnBalance)
	if err != nil {
		return err
	}

	return w.updateAccountBalances(acc)
}

// AccountAddCheckBalance updates ptc, psc and eth balance values.
func (w *Worker) AccountAddCheckBalance(job *data.Job) error {
	acc, err := w.relatedAccount(job, data.JobAccountAddCheckBalance)
	if err != nil {
		// Account was deleted, stop updating.
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	if err = w.updateAccountBalances(acc); err != nil {
		return err
	}

	// HACK: return error to repeat job after a minute.
	job.NotBefore = time.Now().Add(time.Minute)
	return fmt.Errorf("repeating job")
}

func (w *Worker) afterChannelTopUp(job *data.Job, agent bool) error {
	var jobType string

	if agent {
		jobType = data.JobAgentAfterChannelTopUp
	} else {
		jobType = data.JobClientAfterChannelTopUp
	}

	channel, err := w.relatedChannel(job, jobType)
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
		return fmt.Errorf("related channel does" +
			" not correspond to log input")
	}

	channel.TotalDeposit += logInput.addedDeposit.Uint64()
	if err = w.db.Update(channel); err != nil {
		return fmt.Errorf("could not update channels deposit: %v", err)
	}

	return nil
}
