package handler

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/privatix/dappctrl/eth/contract"
)

// EthBackend adapter to communicate with contract.
type EthBackend interface {
	CooperativeClose(*bind.TransactOpts, common.Address, uint32,
		[common.HashLength]byte, *big.Int, []byte, []byte) error

	GetTransactionByHash(context.Context, common.Hash) (*types.Transaction, bool, error)

	RegisterServiceOffering(*bind.TransactOpts, [common.HashLength]byte,
		*big.Int, uint16) error

	PTCBalanceOf(*bind.CallOpts, common.Address) (*big.Int, error)

	PTCIncreaseApproval(*bind.TransactOpts, common.Address, *big.Int) error

	PSCBalanceOf(*bind.CallOpts, common.Address) (*big.Int, error)

	PSCAddBalanceERC20(*bind.TransactOpts, *big.Int) error

	PSCReturnBalanceERC20(*bind.TransactOpts, *big.Int) error
}

type ethBackendInstance struct {
	psc  *contract.PrivatixServiceContract
	ptc  *contract.PrivatixTokenContract
	conn *ethclient.Client
}

// NewEthBackend returns eth back implementation.
func NewEthBackend(psc *contract.PrivatixServiceContract,
	ptc *contract.PrivatixTokenContract, conn *ethclient.Client) EthBackend {
	return &ethBackendInstance{psc, ptc, conn}
}

func (b *ethBackendInstance) CooperativeClose(opts *bind.TransactOpts,
	agent common.Address, block uint32, offeringHash [common.HashLength]byte,
	balance *big.Int, balanceSig, closingSig []byte) error {
	_, err := b.psc.CooperativeClose(opts, agent, block, offeringHash,
		balance, balanceSig, closingSig)
	return err
}

func (b *ethBackendInstance) GetTransactionByHash(ctx context.Context,
	hash common.Hash) (*types.Transaction, bool, error) {
	return b.conn.TransactionByHash(ctx, hash)
}

func (b *ethBackendInstance) RegisterServiceOffering(opts *bind.TransactOpts,
	offeringHash [common.HashLength]byte,
	minDeposit *big.Int, maxSupply uint16) error {
	_, err := b.psc.RegisterServiceOffering(opts, offeringHash,
		minDeposit, maxSupply)
	return err
}

func (b *ethBackendInstance) PTCBalanceOf(opts *bind.CallOpts,
	owner common.Address) (*big.Int, error) {
	return b.ptc.BalanceOf(opts, owner)
}

func (b *ethBackendInstance) PTCIncreaseApproval(opts *bind.TransactOpts,
	spender common.Address, addedVal *big.Int) error {
	_, err := b.ptc.IncreaseApproval(opts, spender, addedVal)
	return err
}

func (b *ethBackendInstance) PSCBalanceOf(opts *bind.CallOpts,
	owner common.Address) (*big.Int, error) {
	return b.psc.BalanceOf(opts, owner)
}

func (b *ethBackendInstance) PSCAddBalanceERC20(opts *bind.TransactOpts,
	amount *big.Int) error {
	_, err := b.psc.AddBalanceERC20(opts, amount)
	return err
}

func (b *ethBackendInstance) PSCReturnBalanceERC20(opts *bind.TransactOpts,
	amount *big.Int) error {
	_, err := b.psc.ReturnBalanceERC20(opts, amount)
	return err
}
