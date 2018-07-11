package worker

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/privatix/dappctrl/eth/contract"
)

// EthBackend adapter to communicate with contract.
type EthBackend interface {
	LatestBlockNumber(ctx context.Context) (*big.Int, error)
	CooperativeClose(*bind.TransactOpts, common.Address, uint32,
		[common.HashLength]byte, *big.Int, []byte, []byte) (*types.Transaction, error)

	GetTransactionByHash(context.Context, common.Hash) (*types.Transaction, bool, error)

	RegisterServiceOffering(*bind.TransactOpts, [common.HashLength]byte,
		*big.Int, uint16) (*types.Transaction, error)

	PTCBalanceOf(*bind.CallOpts, common.Address) (*big.Int, error)

	PTCIncreaseApproval(*bind.TransactOpts, common.Address, *big.Int) (*types.Transaction, error)

	PSCBalanceOf(*bind.CallOpts, common.Address) (*big.Int, error)

	PSCAddBalanceERC20(*bind.TransactOpts, *big.Int) (*types.Transaction, error)

	PSCGetChannelInfo(opts *bind.CallOpts,
		client common.Address, agent common.Address,
		blockNumber uint32,
		hash [common.HashLength]byte) ([common.HashLength]byte,
		*big.Int, uint32, *big.Int, error)

	PSCReturnBalanceERC20(*bind.TransactOpts, *big.Int) (*types.Transaction, error)

	PSCOfferingSupply(opts *bind.CallOpts,
		hash [common.HashLength]byte) (uint16, error)

	PSCCreateChannel(opts *bind.TransactOpts,
		agent common.Address, hash [common.HashLength]byte,
		deposit *big.Int) (*types.Transaction, error)

	PSCTopUpChannel(opts *bind.TransactOpts, agent common.Address,
		blockNumber uint32, hash [common.HashLength]byte,
		deposit *big.Int) (*types.Transaction, error)

	PSCUncooperativeClose(opts *bind.TransactOpts, agent common.Address,
		blockNumber uint32, hash [common.HashLength]byte,
		balance *big.Int) (*types.Transaction, error)

	EthBalanceAt(context.Context, common.Address) (*big.Int, error)

	PSCSettle(opts *bind.TransactOpts,
		agent common.Address, blockNumber uint32,
		hash [common.HashLength]byte) (*types.Transaction, error)
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

func (b *ethBackendInstance) LatestBlockNumber(ctx context.Context) (*big.Int,
	error) {
	header, err := b.conn.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get"+
			" latest block: %s", err)
	}
	return header.Number, err
}

func (b *ethBackendInstance) CooperativeClose(opts *bind.TransactOpts,
	agent common.Address, block uint32, offeringHash [common.HashLength]byte,
	balance *big.Int, balanceSig, closingSig []byte) (*types.Transaction, error) {
	tx, err := b.psc.CooperativeClose(opts, agent, block, offeringHash,
		balance, balanceSig, closingSig)
	if err != nil {
		return nil, fmt.Errorf("failed to do cooperative close: %s", err)
	}
	return tx, nil
}

func (b *ethBackendInstance) GetTransactionByHash(ctx context.Context,
	hash common.Hash) (*types.Transaction, bool, error) {
	tx, pending, err := b.conn.TransactionByHash(ctx, hash)
	if err != nil {
		err = fmt.Errorf("failed to get transaction by hash: %s", err)
	}
	return tx, pending, err
}

func (b *ethBackendInstance) RegisterServiceOffering(opts *bind.TransactOpts,
	offeringHash [common.HashLength]byte,
	minDeposit *big.Int, maxSupply uint16) (*types.Transaction, error) {
	tx, err := b.psc.RegisterServiceOffering(opts, offeringHash,
		minDeposit, maxSupply)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to register service offering: %s", err)
	}
	return tx, nil
}

func (b *ethBackendInstance) PTCBalanceOf(opts *bind.CallOpts,
	owner common.Address) (*big.Int, error) {
	val, err := b.ptc.BalanceOf(opts, owner)
	if err != nil {
		err = fmt.Errorf("failed to get PTC balance: %s", err)
	}
	return val, err
}

func (b *ethBackendInstance) PTCIncreaseApproval(opts *bind.TransactOpts,
	spender common.Address, addedVal *big.Int) (*types.Transaction, error) {
	tx, err := b.ptc.IncreaseApproval(opts, spender, addedVal)
	if err != nil {
		return nil, fmt.Errorf("failed to PTC increase approval: %s", err)
	}
	return tx, nil
}

func (b *ethBackendInstance) PSCBalanceOf(opts *bind.CallOpts,
	owner common.Address) (*big.Int, error) {
	val, err := b.psc.BalanceOf(opts, owner)
	if err != nil {
		err = fmt.Errorf("failed to get PSC balance: %s", err)
	}
	return val, err
}

func (b *ethBackendInstance) PSCAddBalanceERC20(opts *bind.TransactOpts,
	amount *big.Int) (*types.Transaction, error) {
	tx, err := b.psc.AddBalanceERC20(opts, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to add ERC20 balance: %s", err)
	}
	return tx, nil
}

func (b *ethBackendInstance) PSCOfferingSupply(
	opts *bind.CallOpts, hash [common.HashLength]byte) (uint16, error) {
	supply, err := b.psc.GetOfferingSupply(opts, hash)
	if err != nil {
		err = fmt.Errorf("failed to get PSC offering supply: %s", err)
	}
	return supply, err
}

func (b *ethBackendInstance) PSCGetChannelInfo(opts *bind.CallOpts,
	client common.Address, agent common.Address,
	blockNumber uint32,
	hash [common.HashLength]byte) ([common.HashLength]byte,
	*big.Int, uint32, *big.Int, error) {
	return b.psc.GetChannelInfo(opts, client, agent, blockNumber, hash)
}

func (b *ethBackendInstance) PSCCreateChannel(opts *bind.TransactOpts,
	agent common.Address, hash [common.HashLength]byte,
	deposit *big.Int) (*types.Transaction, error) {
	tx, err := b.psc.CreateChannel(opts, agent, hash, deposit, common.Hash{})
	if err != nil {
		err = fmt.Errorf("failed to create PSC channel: %s", err)
	}
	return tx, err
}

func (b *ethBackendInstance) PSCTopUpChannel(opts *bind.TransactOpts,
	agent common.Address, blockNumber uint32, hash [common.HashLength]byte,
	deposit *big.Int) (*types.Transaction, error) {
	tx, err := b.psc.TopUpChannel(opts, agent, blockNumber, hash, deposit)
	if err != nil {
		err = fmt.Errorf("failed to top up PSC channel: %s", err)
	}
	return tx, err
}

func (b *ethBackendInstance) PSCUncooperativeClose(opts *bind.TransactOpts,
	agent common.Address, blockNumber uint32, hash [common.HashLength]byte,
	balance *big.Int) (*types.Transaction, error) {
	tx, err := b.psc.UncooperativeClose(opts, agent,
		blockNumber, hash, balance)
	if err != nil {
		err = fmt.Errorf("failed to uncooperative close"+
			" PSC channel: %s", err)
	}
	return tx, err
}

func (b *ethBackendInstance) PSCReturnBalanceERC20(opts *bind.TransactOpts,
	amount *big.Int) (*types.Transaction, error) {
	tx, err := b.psc.ReturnBalanceERC20(opts, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to return ERC20 balance: %s", err)
	}
	return tx, nil
}

func (b *ethBackendInstance) EthBalanceAt(ctx context.Context,
	owner common.Address) (*big.Int, error) {
	return b.conn.BalanceAt(ctx, owner, nil)
}

func (b *ethBackendInstance) PSCSettle(opts *bind.TransactOpts,
	agent common.Address, blockNumber uint32,
	hash [common.HashLength]byte) (*types.Transaction, error) {
	tx, err := b.psc.Settle(opts, agent, blockNumber, hash)
	if err != nil {
		err = fmt.Errorf("failed to settle"+
			" PSC channel: %s", err)
	}
	return tx, err
}
