package worker

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/privatix/dappctrl/util/log"
)

func (w *Worker) getTransaction(logger log.Logger, hash common.Hash) (*types.Transaction, error) {
	ethTx, pen, err := w.ethBack.GetTransactionByHash(context.Background(), hash)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrEthGetTransaction
	}

	if pen {
		return nil, ErrEthTxInPendingState
	}

	return ethTx, nil
}
