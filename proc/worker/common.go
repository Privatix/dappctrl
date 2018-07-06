package worker

import (
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/privatix/dappctrl/data"
)

// DoPreAccountAddBalanceApprove is an actual logic behind the
// PreAccountAddBalanceApprove handler.
func (w *Worker) DoPreAccountAddBalanceApprove(acc *data.Account,
	gasPrice uint64, amount uint, job *data.Job) (string, error) {
	addr, err := data.ToAddress(acc.EthAddr)
	if err != nil {
		return "", fmt.Errorf("unable to parse account's addr: %v", err)
	}

	balance, err := w.ethBack.PTCBalanceOf(&bind.CallOpts{}, addr)
	if err != nil {
		return "", fmt.Errorf(
			"could not get account's ptc balance: %v", err)
	}

	if balance.Uint64() < uint64(amount) {
		return "", fmt.Errorf("insufficient ptc balance")
	}

	balance, err = w.ethBalance(addr)
	if err != nil {
		return "", fmt.Errorf("failed to get eth balance: %v", err)
	}

	wantedEthBalance := w.gasConf.PTC.Approve * gasPrice

	if wantedEthBalance > balance.Uint64() {
		return "", fmt.Errorf(
			"unsufficient eth balance, wanted: %v, got: %v",
			wantedEthBalance, balance.Uint64())
	}

	key, err := w.key(acc.PrivateKey)
	if err != nil {
		return "", fmt.Errorf(
			"unable to parse account's priv key: %v", err)
	}

	auth := bind.NewKeyedTransactor(key)
	auth.GasLimit = w.gasConf.PTC.Approve
	auth.GasPrice = big.NewInt(int64(gasPrice))
	tx, err := w.ethBack.PTCIncreaseApproval(auth,
		w.pscAddr, big.NewInt(int64(amount)))
	if err != nil {
		return "", fmt.Errorf("could not ptc increase approve: %v", err)
	}

	err = w.saveEthTX(job, tx, "PTCIncreaseApproval", data.JobAccount,
		acc.ID, acc.EthAddr, data.FromBytes(w.pscAddr.Bytes()))
	if err != nil {
		return "", err
	}

	return data.FromBytes(tx.Hash().Bytes()), nil
}

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

	_, err = w.DoPreAccountAddBalanceApprove(
		acc, jobData.GasPrice, jobData.Amount, job)
	return err
}

// PreAccountAddBalance adds balance to psc.
func (w *Worker) PreAccountAddBalance(job *data.Job) error {
	acc, err := w.relatedAccount(job, data.JobPreAccountAddBalance)
	if err != nil {
		return err
	}

	jobData, err := w.approvedBalanceData(job)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %v", err)
	}

	key, err := w.key(acc.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to parse account's priv key: %v", err)
	}

	auth := bind.NewKeyedTransactor(key)
	auth.GasLimit = w.gasConf.PSC.AddBalanceERC20
	auth.GasPrice = big.NewInt(int64(jobData.GasPrice))
	tx, err := w.ethBack.PSCAddBalanceERC20(auth, big.NewInt(int64(jobData.Amount)))
	if err != nil {
		return fmt.Errorf("could not add balance to psc: %v", err)
	}

	return w.saveEthTX(job, tx, "PSCAddBalanceERC20", job.RelatedType,
		job.RelatedID, acc.EthAddr, data.FromBytes(w.pscAddr.Bytes()))
}

func (w *Worker) approvedBalanceData(job *data.Job) (*data.JobBalanceData, error) {
	ethLog, err := w.ethLog(job)
	if err != nil {
		return nil, err
	}
	approveJob := data.Job{}
	err = w.db.SelectOneTo(&approveJob,
		`INNER JOIN eth_txs ON
			jobs.id=eth_txs.job AND
			eth_txs.hash=$1 AND
			jobs.related_type=$2 AND
			jobs.related_id=$3 AND
			jobs.type=$4`,
		ethLog.TxHash, data.JobAccount, job.RelatedID,
		data.JobPreAccountAddBalanceApprove)
	if err != nil {
		return nil, err
	}

	return w.balanceData(&approveJob)
}

// AfterAccountAddBalance updates psc and ptc balance of an account.
func (w *Worker) AfterAccountAddBalance(job *data.Job) error {
	acc, err := w.relatedAccount(job, data.JobAfterAccountAddBalance)
	if err != nil {
		return err
	}

	return w.updateAccountBalances(acc)
}

// DoPreAccountReturnBalance is an actual logic behind the
// PreAccountReturnBalance handler.
func (w *Worker) DoPreAccountReturnBalance(acc *data.Account,
	gasPrice uint64, amount uint, job *data.Job) (string, error) {
	key, err := w.key(acc.PrivateKey)
	if err != nil {
		return "", fmt.Errorf(
			"unable to parse account's priv key: %v", err)
	}

	auth := bind.NewKeyedTransactor(key)

	balance, err := w.ethBack.PSCBalanceOf(&bind.CallOpts{}, auth.From)
	if err != nil {
		return "", fmt.Errorf(
			"could not get account's psc balance: %v", err)
	}

	if balance.Uint64() < uint64(amount) {
		return "", fmt.Errorf("insufficient psc balance")
	}

	balance, err = w.ethBalance(auth.From)
	if err != nil {
		return "", fmt.Errorf("failed to get eth balance: %v", err)
	}

	wantedEthBalance := w.gasConf.PSC.ReturnBalanceERC20 * gasPrice

	if wantedEthBalance > balance.Uint64() {
		return "", fmt.Errorf(
			"unsufficient eth balance, wanted: %v, got: %v",
			wantedEthBalance, balance.Uint64())
	}

	auth.GasLimit = w.gasConf.PSC.ReturnBalanceERC20
	auth.GasPrice = big.NewInt(int64(gasPrice))

	tx, err := w.ethBack.PSCReturnBalanceERC20(
		auth, big.NewInt(int64(amount)))
	if err != nil {
		return "", fmt.Errorf(
			"could not return balance from psc: %v", err)
	}

	err = w.saveEthTX(job, tx, "PSCReturnBalanceERC20", data.JobAccount,
		acc.ID, data.FromBytes(w.pscAddr.Bytes()), acc.EthAddr)
	if err != nil {
		return "", err
	}

	return data.FromBytes(tx.Hash().Bytes()), nil
}

// PreAccountReturnBalance returns from psc to ptc.
func (w *Worker) PreAccountReturnBalance(job *data.Job) error {
	acc, err := w.relatedAccount(job, data.JobPreAccountReturnBalance)
	if err != nil {
		return err
	}

	jobData, err := w.balanceData(job)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %v", err)
	}

	_, err = w.DoPreAccountReturnBalance(
		acc, jobData.GasPrice, jobData.Amount, job)
	return err
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

	// Repeat job after a minute.
	newJob := *job
	newJob.NotBefore = time.Now().Add(time.Minute)
	return w.queue.Add(&newJob)
}

func (w *Worker) afterChannelTopUp(job *data.Job, jobType string) error {
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
