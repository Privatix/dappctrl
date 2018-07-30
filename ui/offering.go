package ui

import (
	"context"

	"github.com/ethereum/go-ethereum/rpc"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/proc/worker"
	"github.com/privatix/dappctrl/util"
)

// AcceptOffering initiates JobClientPreChannelCreate job and subscribes to job
// results of the corresponding flow.
func (h *Handler) AcceptOffering(ctx context.Context,
	password, account, offering string,
	gasPrice uint64) (*rpc.Subscription, error) {
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

	return h.subscribeToJobResults(ctx, logger, rid)
}
