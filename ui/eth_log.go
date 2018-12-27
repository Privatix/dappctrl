package ui

import (
	"database/sql"
	"fmt"
)

// GetLastBlockNumber returns last known block number.
func (h *Handler) GetLastBlockNumber(tkn string) (*uint64, error) {
	logger := h.logger.Add("method", "GetLastBlockNumber")

	if !h.token.Check(tkn) {
		return nil, ErrAccessDenied
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

	logger.Error(fmt.Sprint(queryRet.Int64))
	ret := uint64(queryRet.Int64) + minConfirmations
	return &ret, nil
}
