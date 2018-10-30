package ui

import (
	"fmt"
	"strings"

	"github.com/privatix/dappctrl/data"
)

// AccountAggregatedType is related type to aggregate transactions for
// a specific account.
const AccountAggregatedType = "accountAggregated"

// GetEthTransactionsResult is result of GetEthTransactions method.
type GetEthTransactionsResult struct {
	Items      []data.EthTx `json:"items"`
	TotalItems int          `json:"totalItems"`
}

// GetEthTransactions returns transactions by related object.
func (h *Handler) GetEthTransactions(password, relType, relID string,
	offset, limit uint) (*GetEthTransactionsResult, error) {
	logger := h.logger.Add("method", "GetEthTransactions", "relatedType",
		relType, "relatedID", relID, "limit", limit, "offset", offset)

	if err := h.checkPassword(logger, password); err != nil {
		return nil, err
	}

	conds := make([]string, 0)
	args := make([]interface{}, 0)
	// If the relType is `accountAggregated`, then gets an Ethereum
	// address of the account and find all transactions where this address
	// is the sender.
	if relType == AccountAggregatedType {
		var acc data.Account
		if err := h.findByPrimaryKey(
			logger, ErrAccountNotFound, &acc, relID); err != nil {
			return nil, err
		}
		args = append(args, acc.EthAddr)
		conds = append(
			conds,
			"addr_from="+h.db.Placeholder(len(args)))
	} else if relType != "" {
		args = append(args, relType)
		conds = append(
			conds,
			"related_type="+h.db.Placeholder(len(args)))
	}
	if relID != "" && relType != AccountAggregatedType {
		args = append(args, relID)
		conds = append(
			conds,
			"related_id="+h.db.Placeholder(len(args)))
	}

	tail := ""
	if len(conds) > 0 {
		tail = "WHERE " + strings.Join(conds, " AND ")
	}

	count, err := h.numberOfObjects(
		logger, data.EthTxTable.Name(), tail, args)
	if err != nil {
		return nil, err
	}

	offsetLimit := h.offsetLimit(offset, limit)

	sorting := `ORDER BY issued DESC`

	tail = fmt.Sprintf("%s %s %s", tail, sorting, offsetLimit)

	txs, err := h.selectAllFrom(logger, data.EthTxTable, tail, args...)
	if err != nil {
		return nil, err
	}

	ret := make([]data.EthTx, len(txs))

	for i, v := range txs {
		ret[i] = *v.(*data.EthTx)
	}
	return &GetEthTransactionsResult{ret, count}, nil
}
