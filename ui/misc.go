package ui

import (
	"context"

	"github.com/ethereum/go-ethereum/rpc"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/util/rpcsrv"
)

const (
	activeOfferingCondition = `
		offer_status = 'register'
			AND status = 'msg_channel_published'
			AND NOT is_local
			AND current_supply > 0
			AND agent NOT IN (SELECT eth_addr FROM accounts)`
)

func (h *Handler) findByPrimaryKey(logger log.Logger,
	notFoundError error, record reform.Record, id string) error {
	if err := h.db.FindByPrimaryKeyTo(record, id); err != nil {
		logger.Error(err.Error())
		if err == reform.ErrNoRows {
			return notFoundError
		}
		return ErrInternal
	}
	return nil
}

func (h *Handler) checkPassword(logger log.Logger, password string) error {
	hash, err := data.ReadSetting(h.db.Querier, data.SettingPasswordHash)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	salt, err := data.ReadSetting(h.db.Querier, data.SettingPasswordSalt)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	err = data.ValidatePassword(hash, password, salt)
	if err != nil {
		logger.Error(err.Error())
		return ErrAccessDenied
	}

	return nil
}

func (h *Handler) findActiveOfferingByID(
	logger log.Logger, id string) (*data.Offering, error) {
	var offer data.Offering
	if err := h.db.SelectOneTo(&offer,
		"WHERE id = $1 AND "+activeOfferingCondition, id); err != nil {
		logger.Error(err.Error())
		if err == reform.ErrNoRows {
			return nil, ErrOfferingNotFound
		}
		return nil, ErrInternal
	}
	return &offer, nil
}

// JobResult is a job result notification.
type JobResult struct {
	Type   string        `json:"type"`
	Result *rpcsrv.Error `json:"result"`
}

func (h *Handler) subscribeToJobResults(ctx context.Context,
	logger log.Logger, relatedID string) (*rpc.Subscription, error) {
	ntf, ok := rpc.NotifierFromContext(ctx)
	if !ok {
		logger.Error("no notifier found in context")
		return nil, ErrInternal
	}

	sub := ntf.CreateSubscription()

	sid := string(sub.ID)
	err := h.queue.Subscribe(relatedID, sid,
		func(job *data.Job, result error) {
			ntf.Notify(sub.ID,
				&JobResult{job.Type, rpcsrv.ToError(result)})
		})
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	go func() {
		for {
			if err, _ := <-sub.Err(); err != nil {
				logger.Warn(err.Error())
			}

			err := h.queue.Unsubscribe(relatedID, sid)
			if err != nil {
				logger.Error(err.Error())
			}

			break
		}
	}()

	return sub, nil
}
