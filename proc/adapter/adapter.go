package adapter

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/privatix/dappctrl/eth/contract"
)

const (
	second = time.Second
)

// EthBackend adapter to communicate with contract.
type EthBackend interface {
	LatestBlockNumber(ctx context.Context) (*big.Int, error)

	SuggestGasPrice(ctx context.Context) (*big.Int, error)

	EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error)

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

	PSCGetOfferingInfo(opts *bind.CallOpts,
		hash [common.HashLength]byte) (agentAddr common.Address,
		minDeposit *big.Int, maxSupply uint16, currentSupply uint16,
		updateBlockNumber uint32, active bool, err error)

	PSCGetPopUpPeriod(
		opts *bind.CallOpts) (uint32, error)

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

	PSCRemoveServiceOffering(opts *bind.TransactOpts,
		offeringHash [32]byte) (*types.Transaction, error)

	PSCPopupServiceOffering(opts *bind.TransactOpts,
		offeringHash [32]byte) (*types.Transaction, error)
}

type ethBackendInstance struct {
	psc     *contract.PrivatixServiceContract
	ptc     *contract.PrivatixTokenContract
	conn    *ethclient.Client
	timeout uint64
}

// NewEthBackend returns eth back implementation.
func NewEthBackend(psc *contract.PrivatixServiceContract,
	ptc *contract.PrivatixTokenContract, conn *ethclient.Client,
	timeout uint64) EthBackend {
	return &ethBackendInstance{
		psc:     psc,
		ptc:     ptc,
		conn:    conn,
		timeout: timeout,
	}
}

// AddTimeout adds timeout to context.
func (b *ethBackendInstance) AddTimeout(
	ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx,
		time.Duration(b.timeout)*second)
}

// LatestBlockNumber returns a block number from the current canonical chain.
func (b *ethBackendInstance) LatestBlockNumber(ctx context.Context) (*big.Int,
	error) {
	ctx2, cancel := b.AddTimeout(ctx)
	defer cancel()

	header, err := b.conn.HeaderByNumber(ctx2, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get"+
			" latest block: %s", err)
	}
	return header.Number, err
}

// SuggestGasPrice retrieves the currently suggested gas price to allow a timely
// execution of a transaction.
func (b *ethBackendInstance) SuggestGasPrice(
	ctx context.Context) (*big.Int, error) {
	ctx2, cancel := b.AddTimeout(ctx)
	defer cancel()

	gasPrice, err := b.conn.SuggestGasPrice(ctx2)
	if err != nil {
		return nil, fmt.Errorf("failed to get"+
			" suggested gas price: %s", err)
	}
	return gasPrice, err
}

// EstimateGas tries to estimate the gas needed to execute a specific
// transaction based on the current pending state of the backend blockchain.
func (b *ethBackendInstance) EstimateGas(
	ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	ctx2, cancel := b.AddTimeout(ctx)
	defer cancel()

	gas, err = b.conn.EstimateGas(ctx2, call)
	if err != nil {
		return 0, fmt.Errorf("failed to estimated gas: %s", err)
	}
	return gas, err
}

// CooperativeClose calls cooperativeClose method of Privatix service contract.
func (b *ethBackendInstance) CooperativeClose(opts *bind.TransactOpts,
	agent common.Address, block uint32, offeringHash [common.HashLength]byte,
	balance *big.Int, balanceSig, closingSig []byte) (*types.Transaction, error) {
	ctx2, cancel := b.AddTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	tx, err := b.psc.CooperativeClose(opts, agent, block, offeringHash,
		balance, balanceSig, closingSig)
	if err != nil {
		return nil, fmt.Errorf("failed to do cooperative close: %s", err)
	}
	return tx, nil
}

// TransactionByHash returns the transaction with the given hash.
func (b *ethBackendInstance) GetTransactionByHash(ctx context.Context,
	hash common.Hash) (*types.Transaction, bool, error) {
	ctx2, cancel := b.AddTimeout(ctx)
	defer cancel()

	tx, pending, err := b.conn.TransactionByHash(ctx2, hash)
	if err != nil {
		err = fmt.Errorf("failed to get transaction by hash: %s", err)
	}
	return tx, pending, err
}

// RegisterServiceOffering calls registerServiceOffering method of Privatix
// service contract.
func (b *ethBackendInstance) RegisterServiceOffering(opts *bind.TransactOpts,
	offeringHash [common.HashLength]byte,
	minDeposit *big.Int, maxSupply uint16) (*types.Transaction, error) {
	ctx2, cancel := b.AddTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	tx, err := b.psc.RegisterServiceOffering(opts, offeringHash,
		minDeposit, maxSupply)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to register service offering: %s", err)
	}
	return tx, nil
}

// PTCBalanceOf calls balanceOf method of Privatix token contract.
func (b *ethBackendInstance) PTCBalanceOf(opts *bind.CallOpts,
	owner common.Address) (*big.Int, error) {
	ctx2, cancel := b.AddTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	val, err := b.ptc.BalanceOf(opts, owner)
	if err != nil {
		err = fmt.Errorf("failed to get PTC balance: %s", err)
	}
	return val, err
}

// PTCIncreaseApproval calls increaseApproval method of Privatix token contract.
func (b *ethBackendInstance) PTCIncreaseApproval(opts *bind.TransactOpts,
	spender common.Address, addedVal *big.Int) (*types.Transaction, error) {
	ctx2, cancel := b.AddTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	tx, err := b.ptc.IncreaseApproval(opts, spender, addedVal)
	if err != nil {
		return nil, fmt.Errorf("failed to PTC increase approval: %s", err)
	}
	return tx, nil
}

// PSCBalanceOf calls balanceOf method of Privatix service contract.
func (b *ethBackendInstance) PSCBalanceOf(opts *bind.CallOpts,
	owner common.Address) (*big.Int, error) {
	ctx2, cancel := b.AddTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	val, err := b.psc.BalanceOf(opts, owner)
	if err != nil {
		err = fmt.Errorf("failed to get PSC balance: %s", err)
	}
	return val, err
}

// PSCAddBalanceERC20 calls addBalanceERC20 of Privatix service contract.
func (b *ethBackendInstance) PSCAddBalanceERC20(opts *bind.TransactOpts,
	amount *big.Int) (*types.Transaction, error) {
	ctx2, cancel := b.AddTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	tx, err := b.psc.AddBalanceERC20(opts, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to add ERC20 balance: %s", err)
	}
	return tx, nil
}

// PSCGetOfferingInfo calls getOfferingInfo of Privatix service contract.
func (b *ethBackendInstance) PSCGetOfferingInfo(opts *bind.CallOpts,
	hash [common.HashLength]byte) (agentAddr common.Address,
	minDeposit *big.Int, maxSupply uint16, currentSupply uint16,
	updateBlockNumber uint32, active bool, err error) {
	ctx2, cancel := b.AddTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	agentAddr, minDeposit, maxSupply, currentSupply,
		updateBlockNumber, err = b.psc.GetOfferingInfo(opts, hash)
	active = updateBlockNumber != 0
	if err != nil {
		err = fmt.Errorf("failed to get PSC offering supply: %s", err)
	}
	return agentAddr, minDeposit, maxSupply, currentSupply,
		updateBlockNumber, active, err
}

// PSCGetPopUpPeriod gets popup_period variable from Privatix service contract.
func (b *ethBackendInstance) PSCGetPopUpPeriod(
	opts *bind.CallOpts) (uint32, error) {
	ctx2, cancel := b.AddTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	return b.psc.PopupPeriod(opts)
}

// PSCGetChannelInfo calls getChannelInfo method of Privatix service contract.
func (b *ethBackendInstance) PSCGetChannelInfo(opts *bind.CallOpts,
	client common.Address, agent common.Address,
	blockNumber uint32,
	hash [common.HashLength]byte) ([common.HashLength]byte,
	*big.Int, uint32, *big.Int, error) {
	ctx2, cancel := b.AddTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	return b.psc.GetChannelInfo(opts, client, agent, blockNumber, hash)
}

// PSCCreateChannel calls createChannel method of Privatix service contract.
func (b *ethBackendInstance) PSCCreateChannel(opts *bind.TransactOpts,
	agent common.Address, hash [common.HashLength]byte,
	deposit *big.Int) (*types.Transaction, error) {
	ctx2, cancel := b.AddTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	tx, err := b.psc.CreateChannel(opts, agent, hash, deposit)
	if err != nil {
		err = fmt.Errorf("failed to create PSC channel: %s", err)
	}
	return tx, err
}

// PSCTopUpChannel calls topUpChannel method of Privatix service contract.
func (b *ethBackendInstance) PSCTopUpChannel(opts *bind.TransactOpts,
	agent common.Address, blockNumber uint32, hash [common.HashLength]byte,
	deposit *big.Int) (*types.Transaction, error) {
	ctx2, cancel := b.AddTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	tx, err := b.psc.TopUpChannel(opts, agent, blockNumber, hash, deposit)
	if err != nil {
		err = fmt.Errorf("failed to top up PSC channel: %s", err)
	}
	return tx, err
}

// PSCUncooperativeClose calls uncooperativeClose method of Privatix service
// contract.
func (b *ethBackendInstance) PSCUncooperativeClose(opts *bind.TransactOpts,
	agent common.Address, blockNumber uint32, hash [common.HashLength]byte,
	balance *big.Int) (*types.Transaction, error) {
	ctx2, cancel := b.AddTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	tx, err := b.psc.UncooperativeClose(opts, agent,
		blockNumber, hash, balance)
	if err != nil {
		err = fmt.Errorf("failed to uncooperative close"+
			" PSC channel: %s", err)
	}
	return tx, err
}

// PSCReturnBalanceERC20 calls returnBalanceERC20 method of Privatix service
// contract.
func (b *ethBackendInstance) PSCReturnBalanceERC20(opts *bind.TransactOpts,
	amount *big.Int) (*types.Transaction, error) {
	ctx2, cancel := b.AddTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	tx, err := b.psc.ReturnBalanceERC20(opts, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to return ERC20 balance: %s", err)
	}
	return tx, nil
}

// EthBalanceAt returns the wei balance of the given account.
func (b *ethBackendInstance) EthBalanceAt(ctx context.Context,
	owner common.Address) (*big.Int, error) {
	ctx2, cancel := b.AddTimeout(ctx)
	defer cancel()

	return b.conn.BalanceAt(ctx2, owner, nil)
}

// PSCSettle calls settle method of Privatix service contract.
func (b *ethBackendInstance) PSCSettle(opts *bind.TransactOpts,
	agent common.Address, blockNumber uint32,
	hash [common.HashLength]byte) (*types.Transaction, error) {
	ctx2, cancel := b.AddTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	tx, err := b.psc.Settle(opts, agent, blockNumber, hash)
	if err != nil {
		err = fmt.Errorf("failed to settle"+
			" PSC channel: %s", err)
	}
	return tx, err
}

// PSCRemoveServiceOffering calls removeServiceOffering method of Privatix
// service contract.
func (b *ethBackendInstance) PSCRemoveServiceOffering(opts *bind.TransactOpts,
	offeringHash [32]byte) (*types.Transaction, error) {
	ctx2, cancel := b.AddTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	tx, err := b.psc.RemoveServiceOffering(opts, offeringHash)
	if err != nil {
		err = fmt.Errorf("failed to remove"+
			" service offering: %v", err)
	}
	return tx, err
}

// PSCPopupServiceOffering calls popupServiceOffering method of  Privatix
// service contract.
func (b *ethBackendInstance) PSCPopupServiceOffering(opts *bind.TransactOpts,
	offeringHash [32]byte) (*types.Transaction, error) {
	ctx2, cancel := b.AddTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	tx, err := b.psc.PopupServiceOffering(opts, offeringHash)
	if err != nil {
		err = fmt.Errorf("failed to pop up"+
			" service offering: %v", err)
	}
	return tx, err
}
