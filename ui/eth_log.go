package ui

import (
	"database/sql"
)

// GetLastBlockNumber returns last known block number.
func (h *Handler) GetLastBlockNumber(password string) (*uint64, error) {
	logger := h.logger.Add("method", "GetLastBlockNumber")

	if err := h.checkPassword(logger, password); err != nil {
		return nil, err
	}

	minConfirmations, err := h.minConfirmations(logger)
	if err != nil {
		return nil, err
	}

	var queryRet sql.NullInt64
	row := h.db.QueryRow("SELECT max((data->'ethereumLog'->>'block') :: bigint) from jobs")
	if err := row.Scan(&queryRet); err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	ret := uint64(queryRet.Int64) + minConfirmations
	return &ret, nil
}
