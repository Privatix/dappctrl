package sess

import (
	"context"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/rpc"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
)

// Adapter connection statuses.
const (
	ConnStart = "start"
	ConnStop  = "stop"
)

// ConnChangeResult is an ConnChange notification result.
type ConnChangeResult struct {
	Channel string `json:"channel,omitempty"`
	Status  string `json:"command,omitempty"`
}

func (h *Handler) handleConnChange(product string, logger log.Logger,
	ntf *rpc.Notifier, sub *rpc.Subscription, job *data.Job, closeCh chan struct{}) {
	var ch data.Channel
	err := data.FindByPrimaryKeyTo(h.db.Querier, &ch, job.RelatedID)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	s := ch.ServiceStatus
	if s != data.ServiceActivating && s != data.ServiceSuspending && s != data.ServiceTerminating {
		return
	}

	var offer data.Offering
	err = data.FindByPrimaryKeyTo(h.db.Querier, &offer, ch.Offering)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	if offer.Product != product {
		return
	}

	status := ConnStart
	if job.Type == data.JobClientPreServiceSuspend ||
		job.Type == data.JobAgentPreServiceSuspend ||
		job.Type == data.JobClientPreServiceTerminate ||
		job.Type == data.JobAgentPreServiceTerminate {
		status = ConnStop
	}

	err = ntf.Notify(sub.ID, &ConnChangeResult{ch.ID, status})
	if err != nil {
		logger.Warn(fmt.Sprintf("could not notify: %v", err))
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			close(closeCh)
		}
	}
}

// ConnChange subscribes to changes for adapter connection changes.
func (h *Handler) ConnChange(ctx context.Context,
	product, productPassword string) (*rpc.Subscription, error) {
	logger := h.logger.Add("method", "ConnChange", "product", product)

	logger.Info("subscribing to adapter connection changes")

	_, err := h.checkProductPassword(logger, product, productPassword)
	if err != nil {
		return nil, err
	}

	ntf, ok := rpc.NotifierFromContext(ctx)
	if !ok {
		logger.Error("no notifier found in context")
		return nil, ErrInternal
	}

	sub := ntf.CreateSubscription()
	closeCh := make(chan struct{})
	cb := func(job *data.Job, result error) {
		if result == nil {
			h.handleConnChange(product, logger, ntf, sub, job, closeCh)
		}
	}
	jobTypes := []string{
		data.JobAgentPreEndpointMsgCreate,
		data.JobAgentPreServiceSuspend,
		data.JobClientPreServiceSuspend,
		data.JobAgentPreServiceUnsuspend,
		data.JobClientPreServiceUnsuspend,
		data.JobAgentPreServiceTerminate,
		data.JobClientPreServiceTerminate,
	}

	// TODO: testing. fix it later.
	sid := string(sub.ID)
	if err = h.queue.Subscribe(jobTypes, sid, cb); err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	go func() {
		select {
		case err := <-sub.Err():
			if err != nil {
				logger.Warn(fmt.Sprintf("subscription error: %v", err))
			}
		case <-closeCh:
		}

		err := h.queue.Unsubscribe(jobTypes, sid)
		if err != nil {
			logger.Error(err.Error())
		}
	}()

	return sub, nil
}
