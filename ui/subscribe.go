package ui

import (
	"context"
	"encoding/json"

	"github.com/ethereum/go-ethereum/rpc"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/rpcsrv"
)

// ObjectChangeResult is an ObjectChange notification result.
type ObjectChangeResult struct {
	Object json.RawMessage `json:"object,omitempty"`
	Job    *data.Job       `json:"job"`
	Error  *rpcsrv.Error   `json:"error,omitempty"`
}

var objectChangeTables = map[string]reform.Table{
	data.JobOffering: data.OfferingTable,
	data.JobChannel:  data.ChannelTable,
	data.JobEndpoint: data.EndpointTable,
	data.JobAccount:  data.AccountTable,
}

// ObjectChange subscribes to changes for objects of a given type.
func (h *Handler) ObjectChange(ctx context.Context, password, objectType string,
	objectIDs []string) (*rpc.Subscription, error) {
	logger := h.logger.Add("method", "ObjectChange",
		"objectType", objectType, "objectIDs", objectIDs)

	if err := h.checkPassword(logger, password); err != nil {
		return nil, err
	}

	table, ok := objectChangeTables[objectType]
	if !ok {
		logger.Warn(ErrBadObjectType.Error())
		return nil, ErrBadObjectType
	}

	ntf, ok := rpc.NotifierFromContext(ctx)
	if !ok {
		logger.Error("no notifier found in context")
		return nil, ErrInternal
	}

	sub := ntf.CreateSubscription()
	cb := func(job *data.Job, result error) {
		obj, err := h.db.FindByPrimaryKeyFrom(table, job.RelatedID)
		if err != nil {
			logger.Error(err.Error())
		}

		var odata json.RawMessage
		if obj != nil {
			odata, err = json.Marshal(obj)
			if err != nil {
				logger.Error(err.Error())
			}
		}

		err = ntf.Notify(sub.ID,
			&ObjectChangeResult{odata, job, rpcsrv.ToError(result)})
		if err != nil {
			logger.Warn(err.Error())
		}
	}

	sid := string(sub.ID)
	err := h.queue.Subscribe(objectIDs, sid, cb)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	go func() {
		for err, ok := <-sub.Err(); ok; {
			if err != nil {
				logger.Warn(err.Error())
			}
		}

		err := h.queue.Unsubscribe(objectIDs, sid)
		if err != nil {
			logger.Error(err.Error())
		}
	}()

	return sub, nil
}
