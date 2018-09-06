package ui

import (
	"strconv"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
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

func (h *Handler) jobPublishData(
	logger log.Logger, gasPrice uint64) (*data.JobPublishData, error) {
	ret := &data.JobPublishData{GasPrice: gasPrice}
	if gasPrice == 0 {
		defGasPrice, err := h.defaultGasPrice(logger)
		if err != nil {
			return nil, err
		}
		ret.GasPrice = defGasPrice
	}
	return ret, nil
}

func (h *Handler) defaultGasPrice(logger log.Logger) (uint64, error) {
	gasPriceSettings := &data.Setting{}
	if err := h.findByColumn(logger, ErrDefailtGasPriceNotFound,
		gasPriceSettings, "key", data.SettingDefaultGasPrice); err != nil {
		return 0, err
	}

	val, err := strconv.ParseUint(gasPriceSettings.Value, 10, 64)
	if err != nil {
		logger.Add("error", err).Error("failed to parse default gas price")
		return 0, ErrInternal
	}

	return val, nil
}
