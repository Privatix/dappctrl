package ui

import (
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/proc/worker"
	"github.com/privatix/dappctrl/util"
)

// AcceptOfferingResult is an AcceptOffering result.
type AcceptOfferingResult struct {
	Channel string `json:"channel"`
}

// AcceptOffering initiates JobClientPreChannelCreate job and subscribes to job
// results of the corresponding flow.
func (h *Handler) AcceptOffering(password, account, offering string,
	gasPrice uint64) (*AcceptOfferingResult, error) {
	logger := h.logger.Add("method", "AcceptOffering",
		"account", account, "offering", offering, "gasPrice", gasPrice)

	if err := h.checkPassword(logger, password); err != nil {
		return nil, err
	}

	var acc data.Account
	if err := h.findByPrimaryKey(
		logger, ErrAccountNotFound, &acc, account); err != nil {
		return nil, err
	}

	if _, err := h.findActiveOfferingByID(logger, offering); err != nil {
		return nil, err
	}

	rid := util.NewUUID()
	jdata := &worker.ClientPreChannelCreateData{
		Account: account, Offering: offering, GasPrice: gasPrice}
	if err := job.AddWithData(h.queue, data.JobClientPreChannelCreate,
		data.JobChannel, rid, data.JobUser, jdata); err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	return &AcceptOfferingResult{rid}, nil
}
