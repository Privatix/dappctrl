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
}

var objectWithHashTypes = map[string]reform.Table{
	TypeTemplate: data.TemplateTable,
	TypeOffering: data.OfferingTable,
	TypeEndpoint: data.EndpointTable,
	TypeEthTx:    data.EthTxTable,
}

// GetObject finds object in a database by id,
// then returns an object on raw JSON format.
func (h *Handler) GetObject(
	tkn, objectType, id string) (json.RawMessage, error) {
	logger := h.logger.Add("method", "GetObject",
		"type", objectType, "id", id)

	if !h.token.Check(tkn) {
		return nil, ErrAccessDenied
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

// GetObjectByHash finds object in a database by hash,
// then returns an object on raw JSON format.
func (h *Handler) GetObjectByHash(
	tkn, objectType, hash string) (json.RawMessage, error) {
	logger := h.logger.Add("method", "GetObjectByHash",
		"type", objectType, "hash", hash)

	if !h.token.Check(tkn) {
		return nil, ErrAccessDenied
	}

	table, ok := objectWithHashTypes[objectType]
	if !ok {
		logger.Warn(ErrBadObjectType.Error())
		return nil, ErrBadObjectType
	}

	obj, err := h.db.FindOneFrom(table, "hash", hash)
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
