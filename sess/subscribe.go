package sess

import (
	"context"

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
	ntf *rpc.Notifier, sub *rpc.Subscription, job *data.Job) {
	var ch data.Channel
	err := data.FindByPrimaryKeyTo(h.db.Querier, &ch, job.RelatedID)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	if ch.ServiceStatus != data.ServiceActivating &&
		ch.ServiceStatus != data.ServiceSuspending &&
		ch.ServiceStatus != data.ServiceTerminating {
		return
	}

	var offer data.Offering
	err = data.FindByPrimaryKeyTo(h.db.Querier, &offer, ch.Offering)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	status := ConnStop
	if job.Type == data.JobClientPreServiceUnsuspend {
		status = ConnStart
	}

	if offer.Product == product {
		err = ntf.Notify(sub.ID, &ConnChangeResult{ch.ID, status})
		if err != nil {
			logger.Warn(err.Error())
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
	cb := func(job *data.Job, result error) {
		if result == nil {
			h.handleConnChange(product, logger, ntf, sub, job)
		}
	}

	jobTypes := []string{
		data.JobClientPreServiceUnsuspend,
		data.JobClientPreServiceSuspend,
		data.JobClientPreServiceTerminate,
	}
	sid := string(sub.ID)
	if err = h.queue.Subscribe(jobTypes, sid, cb); err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	go func() {
		for err, ok := <-sub.Err(); ok; {
			if err != nil {
				logger.Warn(err.Error())
			}
		}

		err := h.queue.Unsubscribe(jobTypes, sid)
		if err != nil {
			logger.Error(err.Error())
		}
	}()

	return sub, nil
}
