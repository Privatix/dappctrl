package ui

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"database/sql"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/errors"
	"github.com/privatix/dappctrl/util/log"
)

const (
	activeOfferingCondition = `
		status in ('registered', 'popped_up')
			AND NOT is_local
			AND current_supply > 0
			AND agent NOT IN (SELECT eth_addr FROM accounts)`

	activeOfferingSorting = `
		      ORDER BY block_number_updated DESC
`
)

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

	err = data.ValidatePassword(data.Base64String(hash), password, salt)
	if err != nil {
		logger.Error(err.Error())
		return ErrAccessDenied
	}

	h.pwdStorage.Set(password)

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
		defGasPrice, err := h.suggestedGasPrice(logger)
		if err != nil {
			return nil, err
		}
		ret.GasPrice = defGasPrice
	}
	return ret, nil
}

func (h *Handler) suggestedGasPrice(logger log.Logger) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	price, err := h.suggestor.SuggestGasPrice(ctx)
	if err != nil {
		logger.Warn(fmt.Sprintf("could not get suggested gas price: %v", err))
		return 0, err
	}
	return price.Uint64(), nil
}

func (h *Handler) minConfirmations(logger log.Logger) (uint64, error) {
	return h.findSettingsUint64(logger, data.SettingMinConfirmations)
}

func (h *Handler) findSettingsUint64(logger log.Logger, key string) (uint64, error) {
	logger = logger.Add("settingKey", key)

	gasPriceSettings, err := h.findSetting(logger, key)
	if err != nil {
		return 0, err
	}

	val, err := strconv.ParseUint(gasPriceSettings.Value, 10, 64)
	if err != nil {
		logger.Add("error", err).Error("failed to parse setting value")
		return 0, ErrInternal
	}

	return val, nil
}

func (h *Handler) findSetting(logger log.Logger, key string) (*data.Setting, error) {
	setting := &data.Setting{}
	if err := h.findByColumn(logger, ErrMinConfirmationsNotFound,
		setting, "key", key); err != nil {
		return nil, err
	}
	return setting, nil
}

func (h *Handler) catchError(logger log.Logger, err error) error {
	if e, ok := err.(errors.Error); ok {
		return e
	}
	logger.Error(err.Error())
	return ErrInternal
}

func (h *Handler) uintFromQuery(logger log.Logger, tkn,
	query string, arg ...interface{}) (*uint, error) {
	if !h.token.Check(tkn) {
		logger.Warn("access denied")
		return nil, ErrAccessDenied
	}

	var queryRet sql.NullInt64
	err := h.db.QueryRow(query, arg...).Scan(&queryRet)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	ret := uint(queryRet.Int64)
	return &ret, nil
}

func (h *Handler) numberOfObjects(logger log.Logger, table, conditions string,
	arguments []interface{}) (count int, err error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table, conditions)
	err = h.db.QueryRow(query, arguments...).Scan(&count)
	if err != nil {
		logger.Error(err.Error())
		return 0, ErrInternal
	}
	return count, err
}

func (h *Handler) offsetLimit(offset, limit uint) string {
	var limitCondition string

	if limit != 0 {
		limitCondition = fmt.Sprintf("LIMIT %d", limit)
	}

	offsetCondition := fmt.Sprintf("OFFSET %d", offset)

	return fmt.Sprintf("%s %s", offsetCondition, limitCondition)
}
