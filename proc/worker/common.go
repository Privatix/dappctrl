package worker

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/privatix/dappctrl/data"
)

// CompleteServiceTransition is an end step of service status transitioning.
func (w *Worker) CompleteServiceTransition(job *data.Job) error {
	logger := w.logger.Add("method", "CompleteServiceTransition",
		"job", job)

	ch, err := w.relatedChannel(
		logger, job, data.JobCompleteServiceTransition)
	if err != nil {
		return err
	}

	logger = logger.Add("channel", ch)

	err = w.unmarshalDataTo(logger, job.Data, &ch.ServiceStatus)
	if err != nil {
		return err
	}

	return w.saveRecord(logger, w.db.Querier, ch)
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

	channel.TotalDeposit += logInput.addedDeposit
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

// IncreaseTxGasPrice resent related_id tranaction with increased gas price.
func (w *Worker) IncreaseTxGasPrice(job *data.Job) error {
	logger := w.logger.Add("method", "IncreaseTxGasPrice", "job", job)

	ethTx, err := w.relatedEthTx(logger, job, data.JobIncreaseTxGasPrice)
	if err != nil {
		return err
	}

	jdata, err := w.publishData(logger, job)
	if err != nil {
		return err
	}

	if jdata.GasPrice <= ethTx.GasPrice {
		return ErrTxNoGasIncrease
	}

	txHash, err := data.HexToHash(ethTx.Hash)
	if err != nil {
		logger.Error(fmt.Sprintf("could not get tx hash from hex representation: %v", err))
	}
	_, pending, err := w.ethBack.GetTransactionByHash(context.Background(), txHash)
	if err != nil {
		logger.Error(fmt.Sprintf("could not get tx by hash: %v", err))
		return ErrEthGetTransaction
	}

	if !pending {
		return ErrEthTxIsMined
	}

	tx := new(types.Transaction)
	if err := tx.UnmarshalJSON(ethTx.TxRaw); err != nil {
		return fmt.Errorf("could not build transaction to send: %v", err)
	}
	tx = types.NewTransaction(tx.Nonce(), *tx.To(), tx.Value(), tx.Gas(),
		new(big.Int).SetUint64(jdata.GasPrice), tx.Data())

	if err := w.ethBack.SendTransaction(context.Background(), tx); err != nil {
		return fmt.Errorf("could not send transaction: %v", err)
	}

	return w.saveEthTX(logger, job, tx, ethTx.Method, data.JobTransaction, ethTx.ID,
		ethTx.AddrFrom, ethTx.AddrTo)
}
