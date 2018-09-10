package ui

import (
	"github.com/privatix/dappctrl/util/log"
	reform "gopkg.in/reform.v1"
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

func (h *Handler) findByColumn(logger log.Logger, notFoundError error,
	record reform.Record, column string, arg interface{}) error {
	if err := h.db.FindOneTo(record, column, arg); err != nil {
		logger.Error(err.Error())
		if err == reform.ErrNoRows {
			return notFoundError
		}
		return ErrInternal
	}
	return nil
}

func beginTX(logger log.Logger, db *reform.DB) (*reform.TX, error) {
	tx, err := db.Begin()
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}
	return tx, nil
}

func commitTX(logger log.Logger, tx *reform.TX) error {
	err := tx.Commit()
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}
	return nil
}

func insert(logger log.Logger, db *reform.Querier, items ...reform.Struct) error {
	for _, item := range items {
		if err := db.Insert(item); err != nil {
			logger.Error(err.Error())
			return ErrInternal
		}
	}
	return nil
}

func update(logger log.Logger, db *reform.Querier, items ...reform.Record) error {
	for _, item := range items {
		if err := db.Update(item); err != nil {
			logger.Error(err.Error())
			return ErrInternal
		}
	}
	return nil
}
