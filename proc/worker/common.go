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
		return fmt.Errorf("insufficient ptc balance")
	}

	amount, err = w.ethBalance(addr)
	if err != nil {
		return fmt.Errorf("failed to get eth balance: %v", err)
	}

	wantedEthBalance := w.gasConf.PTC.Approve * jobData.GasPrice

	if wantedEthBalance > amount.Uint64() {
		return fmt.Errorf("unsufficient eth balance, wanted: %v, got: %v",
			wantedEthBalance, amount.Uint64())
	}

	key, err := w.key(acc.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to parse account's priv key: %v", err)
	}

	auth := bind.NewKeyedTransactor(key)
	auth.GasLimit = w.gasConf.PTC.Approve
	auth.GasPrice = big.NewInt(int64(jobData.GasPrice))
	tx, err := w.ethBack.PTCIncreaseApproval(auth,
		w.pscAddr, big.NewInt(int64(jobData.Amount)))
	if err != nil {
		return fmt.Errorf("could not ptc increase approve: %v", err)
	}

	return w.saveEthTX(job, tx, "PTCIncreaseApproval", job.RelatedType,
		job.RelatedID, acc.EthAddr, data.FromBytes(w.pscAddr.Bytes()))
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
	auth.GasLimit = w.gasConf.PSC.AddBalanceERC20
	tx, err := w.ethBack.PSCAddBalanceERC20(auth, big.NewInt(int64(jobData.Amount)))
	if err != nil {
		return fmt.Errorf("could not add balance to psc: %v", err)
	}

	return w.saveEthTX(job, tx, "PSCAddBalanceERC20", job.RelatedType,
		job.RelatedID, acc.EthAddr, data.FromBytes(w.pscAddr.Bytes()))
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

	jobData, err := w.balanceData(job)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %v", err)
	}

	auth := bind.NewKeyedTransactor(key)

	amount, err := w.ethBack.PSCBalanceOf(&bind.CallOpts{}, auth.From)
	if err != nil {
		return fmt.Errorf("could not get account's psc balance: %v", err)
	}

	if amount.Uint64() < uint64(jobData.Amount) {
		return fmt.Errorf("insufficient psc balance")
	}

	amount, err = w.ethBalance(auth.From)
	if err != nil {
		return fmt.Errorf("failed to get eth balance: %v", err)
	}

	wantedEthBalance := w.gasConf.PSC.ReturnBalanceERC20 * jobData.GasPrice

	if wantedEthBalance > amount.Uint64() {
		return fmt.Errorf("unsufficient eth balance, wanted: %v, got: %v",
			wantedEthBalance, amount.Uint64())
	}

	auth.GasLimit = w.gasConf.PSC.ReturnBalanceERC20
	auth.GasPrice = big.NewInt(int64(jobData.GasPrice))

	tx, err := w.ethBack.PSCReturnBalanceERC20(auth,
		big.NewInt(int64(jobData.Amount)))
	if err != nil {
		return fmt.Errorf("could not return balance from psc: %v", err)
	}

	return w.saveEthTX(job, tx, "PSCReturnBalanceERC20", job.RelatedType,
		job.RelatedID, data.FromBytes(w.pscAddr.Bytes()), acc.EthAddr)
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
