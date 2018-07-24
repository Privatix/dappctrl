package worker

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (w *Worker) getTransaction(hash common.Hash) (*types.Transaction, error) {
	ethTx, pen, err := w.ethBack.GetTransactionByHash(context.Background(), hash)
	if err != nil {
		return nil, err
	}

	if pen {
		return nil, fmt.Errorf("unexpected pending state" +
			" of transaction")
	}

	return ethTx, nil
}
