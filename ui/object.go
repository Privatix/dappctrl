package ui

import (
	"encoding/json"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
)

// Objects types.
const (
	TypeAccount  = "account"
	TypeUser     = "user"
	TypeTemplate = "template"
	TypeProduct  = "product"
	TypeOffering = "offering"
	TypeChannel  = "channel"
	TypeSession  = "session"
	TypeContract = "contract"
	TypeEndpoint = "endpoint"
	TypeJob      = "job"
	TypeEthTx    = "ethTx"
	TypeEthLog   = "ethLog"
)

var objectTypes = map[string]reform.Table{
	TypeAccount:  data.AccountTable,
	TypeUser:     data.UserTable,
	TypeTemplate: data.TemplateTable,
	TypeProduct:  data.ProductTable,
	TypeOffering: data.OfferingTable,
	TypeChannel:  data.ChannelTable,
	TypeSession:  data.SessionTable,
	TypeContract: data.ContractTable,
	TypeEndpoint: data.EndpointTable,
	TypeJob:      data.JobTable,
	TypeEthTx:    data.EthTxTable,
	TypeEthLog:   data.EthLogTable,
}

// GetObject finds object in a database by id,
// then returns an object on raw JSON format.
func (h *Handler) GetObject(
	password, objectType, id string) (json.RawMessage, error) {
	logger := h.logger.Add("method", "GetObject",
		"type", objectType, "id", id)

	err := h.checkPassword(logger, password)
	if err != nil {
		return nil, err
	}

	table, ok := objectTypes[objectType]
	if !ok {
		logger.Warn(ErrBadObjectType.Error())
		return nil, ErrBadObjectType
	}

	obj, err := h.db.FindByPrimaryKeyFrom(table, id)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrObjectNotFound
	}

	raw, err := json.Marshal(obj)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	return raw, nil
}

func (h *Handler) insertObject(object reform.Struct) error {
	logger := h.logger.Add("method", "insertObject", "object", object)

	if err := h.db.Insert(object); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}
	return nil
}
