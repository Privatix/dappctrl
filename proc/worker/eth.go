package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (w *Worker) getTransaction(hash common.Hash) (*types.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(
			w.ethConfig.Timeout.ResponseHeaderTimeout)*
			time.Second)
	defer cancel()

	ethTx, pen, err := w.ethBack.GetTransactionByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	if pen {
		return nil, fmt.Errorf("unexpected pending state" +
			" of transaction")
	}

	return ethTx, nil
}
