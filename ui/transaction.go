package ui

import (
	"strings"

	"github.com/privatix/dappctrl/data"
)

// GetTransactions returns transactions by related object.
func (h *Handler) GetTransactions(
	password, relType, relID string) ([]data.EthTx, error) {
	logger := h.logger.Add("method", "GetTransactions",
		"relatedType", relType, "relatedID", relID)

	if err := h.checkPassword(logger, password); err != nil {
		return nil, err
	}

	conds := make([]string, 0)
	args := make([]interface{}, 0)
	if relType != "" {
		args = append(args, relType)
		conds = append(
			conds,
			"related_type="+h.db.Placeholder(len(args)))
	}
	if relID != "" {
		args = append(args, relID)
		conds = append(
			conds,
			"related_id="+h.db.Placeholder(len(args)))
	}

	tail := ""
	if len(conds) > 0 {
		tail = "WHERE " + strings.Join(conds, " AND ")
	}

	txs, err := h.selectAllFrom(logger, data.EthTxTable, tail, args...)
	if err != nil {
		return nil, err
	}

	ret := make([]data.EthTx, len(txs))

	for i, v := range txs {
		ret[i] = *v.(*data.EthTx)
	}
	return ret, nil
}
