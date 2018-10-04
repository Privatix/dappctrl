package ui

import (
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
)

// TopUpChannel initiates JobClientPreChannelTopUp job.
func (h *Handler) TopUpChannel(password, channel string, gasPrice uint64) error {
	logger := h.logger.Add("method", "TopUpChannel",
		"channel", channel, "gasPrice", gasPrice)

	if err := h.checkPassword(logger, password); err != nil {
		return err
	}

	ch := &data.Channel{}
	if err := h.findByPrimaryKey(logger,
		ErrChannelNotFound, ch, channel); err != nil {
		return err
	}

	jdata, err := h.jobPublishData(logger, gasPrice)
	if err != nil {
		return err
	}

	return job.AddWithData(h.queue, nil, data.JobClientPreChannelTopUp,
		data.JobChannel, ch.ID, data.JobUser, jdata)
}
