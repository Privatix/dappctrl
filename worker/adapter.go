package worker

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// EthBackend adapter to communicate with contract.
type EthBackend interface {
	CooperativeClose(*bind.TransactOpts, common.Address, uint32,
		[common.HashLength]byte, *big.Int, []byte, []byte) error

	GetTransactionByHash(context.Context, common.Hash) (*types.Transaction, bool, error)

	RegisterServiceOffering(*bind.TransactOpts, [common.HashLength]byte,
		*big.Int, uint16) error

	PTCBalanceOf(*bind.CallOpts, common.Address) (*big.Int, error)

	PTCApprove(*bind.TransactOpts, common.Address, *big.Int) error

	PSCBalanceOf(*bind.CallOpts, common.Address) (*big.Int, error)

	PSCAddBalanceERC20(*bind.TransactOpts, *big.Int) error

	PSCReturnBalanceERC20(*bind.TransactOpts, *big.Int) error
}
