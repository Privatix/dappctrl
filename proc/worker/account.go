package worker

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
)

// PreAccountAddBalanceApprove approve balance if amount exists.
func (w *Worker) PreAccountAddBalanceApprove(job *data.Job) error {
	logger := w.logger.Add("method", "PreAccountAddBalanceApprove", "job", job)

	acc, err := w.relatedAccount(logger, job,
		data.JobPreAccountAddBalanceApprove)
	if err != nil {
		return err
	}

	logger = logger.Add("account", acc)

	jobData, err := w.balanceData(logger, job)
	if err != nil {
		return fmt.Errorf("failed to parse job data: %v", err)
	}

	addr, err := data.HexToAddress(acc.EthAddr)
	if err != nil {
		logger.Error(err.Error())
		return ErrParseEthAddr
	}

	amount, err := w.ethBack.PTCBalanceOf(&bind.CallOpts{}, addr)
	if err != nil {
		logger.Error(err.Error())
		return ErrPTCRetrieveBalance
	}

	if amount.Uint64() < uint64(jobData.Amount) {
		return ErrInsufficientPTCBalance
	}

	ethBalance, err := w.ethBalance(logger, addr)
	if err != nil {
		return err
	}

	wantedEthBalance := w.gasConf.PTC.Approve * jobData.GasPrice

	if wantedEthBalance > ethBalance.Uint64() {
		return ErrInsufficientEthBalance
	}

	key, err := w.key(logger, acc.PrivateKey)
	if err != nil {
		return err
	}

	auth := bind.NewKeyedTransactor(key)
	auth.GasLimit = w.gasConf.PTC.Approve
	auth.GasPrice = new(big.Int).SetUint64(jobData.GasPrice)
	tx, err := w.ethBack.PTCIncreaseApproval(auth,
		w.pscAddr, new(big.Int).SetUint64(jobData.Amount))
	if err != nil {
		logger.Error(err.Error())
		return ErrPTCIncreaseApproval
	}

	return w.saveEthTX(logger, job, tx, "PTCIncreaseApproval", job.RelatedType,
		job.RelatedID, acc.EthAddr, data.HexFromBytes(w.pscAddr.Bytes()))
}

// PreAccountAddBalance adds balance to psc.
func (w *Worker) PreAccountAddBalance(job *data.Job) error {
	logger := w.logger.Add("method", "PreAccountAddBalance", "job", job)

	acc, err := w.relatedAccount(logger, job, data.JobPreAccountAddBalance)
	if err != nil {
		return err
	}

	jobData, err := w.approvedBalanceData(logger, job)
	if err != nil {
		return err
	}

	key, err := w.key(logger, acc.PrivateKey)
	if err != nil {
		return err
	}

	auth := bind.NewKeyedTransactor(key)
	auth.GasLimit = w.gasConf.PSC.AddBalanceERC20
	auth.GasPrice = new(big.Int).SetUint64(jobData.GasPrice)
	tx, err := w.ethBack.PSCAddBalanceERC20(
		auth, new(big.Int).SetUint64(jobData.Amount))
	if err != nil {
		logger.Error(err.Error())
		return ErrPSCAddBalance
	}

	return w.saveEthTX(logger, job, tx, "PSCAddBalanceERC20", job.RelatedType,
		job.RelatedID, acc.EthAddr, data.HexFromBytes(w.pscAddr.Bytes()))
}

func (w *Worker) approvedBalanceData(logger log.Logger,
	job *data.Job) (*data.JobBalanceData, error) {
	ethLog, err := w.ethLog(logger, job)
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
		logger.Error(err.Error())
		return nil, ErrFindApprovalBalanceData
	}

	return w.balanceData(logger, &approveJob)
}

// AfterAccountAddBalance updates psc and ptc balance of an account.
func (w *Worker) AfterAccountAddBalance(job *data.Job) error {
	return w.updateAccountBalancesJob(job, data.JobAfterAccountAddBalance)
}

// AfterAccountReturnBalance updates psc and ptc balance of an account.
func (w *Worker) AfterAccountReturnBalance(job *data.Job) error {
	return w.updateAccountBalancesJob(job, data.JobAfterAccountReturnBalance)
}

// AccountUpdateBalances updates ptc, psc and eth balance values.
func (w *Worker) AccountUpdateBalances(job *data.Job) error {
	return w.updateAccountBalancesJob(job, data.JobAccountUpdateBalances)
}

// PreAccountReturnBalance returns from psc to ptc.
func (w *Worker) PreAccountReturnBalance(job *data.Job) error {
	logger := w.logger.Add("method", "PreAccountReturnBalance", "job", job)
	acc, err := w.relatedAccount(logger, job, data.JobPreAccountReturnBalance)
	if err != nil {
		return err
	}

	key, err := w.key(logger, acc.PrivateKey)
	if err != nil {
		return err
	}

	jobData, err := w.balanceData(logger, job)
	if err != nil {
		return err
	}

	auth := bind.NewKeyedTransactor(key)

	amount, err := w.ethBack.PSCBalanceOf(&bind.CallOpts{}, auth.From)
	if err != nil {
		logger.Error(err.Error())
		return ErrPSCReturnBalance
	}

	if amount.Uint64() < uint64(jobData.Amount) {
		return ErrInsufficientPSCBalance
	}

	ethAmount, err := w.ethBalance(logger, auth.From)
	if err != nil {
		return err
	}

	wantedEthBalance := w.gasConf.PSC.ReturnBalanceERC20 * jobData.GasPrice

	if wantedEthBalance > ethAmount.Uint64() {
		return ErrInsufficientEthBalance
	}

	auth.GasLimit = w.gasConf.PSC.ReturnBalanceERC20
	auth.GasPrice = new(big.Int).SetUint64(jobData.GasPrice)

	tx, err := w.ethBack.PSCReturnBalanceERC20(auth,
		new(big.Int).SetUint64(jobData.Amount))
	if err != nil {
		logger.Add("GasLimit", auth.GasLimit,
			"GasPrice", auth.GasPrice).Error(err.Error())
		return ErrPSCRetrieveBalance
	}

	return w.saveEthTX(logger, job, tx, "PSCReturnBalanceERC20", job.RelatedType,
		job.RelatedID, data.HexFromBytes(w.pscAddr.Bytes()), acc.EthAddr)
}

func (w *Worker) afterChannelTopUp(job *data.Job, jobType string) error {
	logger := w.logger.Add("method", "afterChannelTopUp", "job", job)

	channel, err := w.relatedChannel(logger, job, jobType)
	if err != nil {
		return err
	}

	ethLog, err := w.ethLog(logger, job)
	if err != nil {
		return err
	}

	logInput, err := extractLogChannelToppedUp(logger, ethLog)
	if err != nil {
		return fmt.Errorf("could not parse log: %v", err)
	}

	agentAddr, err := data.HexToAddress(channel.Agent)
	if err != nil {
		logger.Error(err.Error())
		return ErrParseEthAddr
	}

	clientAddr, err := data.HexToAddress(channel.Client)
	if err != nil {
		logger.Error(err.Error())
		return ErrParseEthAddr
	}

	logger = logger.Add("agent", agentAddr, "client", clientAddr)

	offering, err := w.offering(logger, channel.Offering)
	if err != nil {
		return err
	}

	offeringHash, err := w.toOfferingHashArr(logger, offering.Hash)
	if err != nil {
		return err
	}

	if agentAddr != logInput.agentAddr ||
		clientAddr != logInput.clientAddr ||
		offeringHash != logInput.offeringHash ||
		channel.Block != logInput.openBlockNum {
		return ErrEthLogChannelMismatch
	}

	channel.TotalDeposit += logInput.addedDeposit.Uint64()
	if err = w.db.Update(channel); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	if job.Type == data.JobClientAfterChannelTopUp {
		account, err := w.account(logger, channel.Client)
		if err != nil {
			return err
		}

		return w.addJob(logger, nil, data.JobAccountUpdateBalances,
			data.JobAccount, account.ID)
	}

	return nil
}

func (w *Worker) findOffering(logger log.Logger, job *data.Job,
	jobType string) (*data.Offering, error) {
	if w.isJobInvalid(job, jobType, data.JobOffering) {
		return nil, ErrInvalidJob
	}

	rec := &data.Offering{}
	err := data.FindByPrimaryKeyTo(w.db.Querier, rec, job.RelatedID)
	if err != nil {
		logger.Warn("offering not found, error: " + err.Error())
		return nil, ErrOfferingNotFound
	}

	return rec, err
}

// DecrementCurrentSupply finds offering and decrements its current supply.
func (w *Worker) DecrementCurrentSupply(job *data.Job) error {
	logger := w.logger.Add("method", "DecrementCurrentSupply", "job", job)

	offering, err := w.findOffering(
		logger, job, data.JobDecrementCurrentSupply)
	if err != nil {
		// Ignore errors if the offering is not found in the database.
		if err == ErrOfferingNotFound {
			return nil
		}
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

	offering, err := w.findOffering(
		logger, job, data.JobIncrementCurrentSupply)
	if err != nil {
		// Ignore errors if the offering is not found in the database.
		if err == ErrOfferingNotFound {
			return nil
		}
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
