package ui

import (
	"fmt"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/util/log"
)

func (h *Handler) tailElements(conditions []string) (elements []string) {
	for k, v := range conditions {
		condition := fmt.Sprintf(
			"%s = %s", v, h.db.Placeholder(k+1))
		elements = append(elements, condition)
	}
	return elements
}

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

func (h *Handler) findAllFrom(logger log.Logger, view reform.View,
	column string, args ...interface{}) ([]reform.Struct, error) {
	rows, err := h.db.FindAllFrom(view, column, args...)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}
	return rows, err
}

func (h *Handler) selectAllFrom(logger log.Logger, view reform.View,
	tail string, args ...interface{}) ([]reform.Struct, error) {
	rows, err := h.db.SelectAllFrom(view, tail, args...)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}
	return rows, err
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
