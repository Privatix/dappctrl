package ui

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

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

func recordByRelatedIDCB(t reform.Table) func(*reform.DB, *data.Job) (reform.Record, error) {
	return func(db *reform.DB, j *data.Job) (reform.Record, error) {
		return db.FindByPrimaryKeyFrom(t, j.RelatedID)
	}
}

var getRecordFuncs = map[string]func(*reform.DB, *data.Job) (reform.Record, error){
	data.JobOffering: recordByRelatedIDCB(data.OfferingTable),
	data.JobChannel:  recordByRelatedIDCB(data.ChannelTable),
	data.JobEndpoint: recordByRelatedIDCB(data.EndpointTable),
	data.JobAccount:  recordByRelatedIDCB(data.AccountTable),
	data.JobTransaction: func(db *reform.DB, j *data.Job) (reform.Record, error) {
		var tx data.EthTx
		err := db.FindOneTo(&tx, "job", j.ID)
		return &tx, err
	},
}

// ObjectChange subscribes to changes for objects of a given type.
func (h *Handler) ObjectChange(ctx context.Context, tkn, objectType string,
	objectIDs []string) (*rpc.Subscription, error) {
	logger := h.logger.Add("method", "ObjectChange",
		"objectType", objectType, "objectIDs", objectIDs)

	if !h.token.Check(tkn) {
		logger.Warn("access denied")
		return nil, ErrAccessDenied
	}

	getRecord, ok := getRecordFuncs[objectType]
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
	sid := string(sub.ID)
	closech := make(chan struct{})
	cb := func(job *data.Job, result error) {
		obj, err := getRecord(h.db, job)
		if err != nil {
			if err == sql.ErrNoRows {
				logger.Add("job", job, "result", result).Warn(err.Error())
			} else {
				logger.Error(err.Error())
			}
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
			logger.Warn(fmt.Sprintf("could not notify subscriber: %v", err))
			close(closech)
		}
	}

	err := h.queue.Subscribe(objectIDs, sid, cb)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	go func() {
		select {
		case err := <-sub.Err():
			if err != nil {
				logger.Warn(fmt.Sprintf("subscription error: %v", err))
			}
		case <-closech:
		}

		err := h.queue.Unsubscribe(objectIDs, sid)
		if err != nil {
			logger.Error(fmt.Sprintf("could not unsubscribe: %v", err))
		}
	}()

	return sub, nil
}
