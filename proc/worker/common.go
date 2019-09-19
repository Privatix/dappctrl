package worker

import (
	"fmt"

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
