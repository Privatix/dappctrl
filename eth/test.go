// +build !notest

package eth

import (
	"context"
	"math/big"
	"reflect"
	"testing"

	"github.com/privatix/dappctrl/data"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TestEthBackCall is mock for Ethereum backend.
type TestEthBackCall struct {
	txOpts *bind.TransactOpts
	method string
	caller common.Address
	args   []interface{}
}

// TestEthBackend is mock for Backend.
type TestEthBackend struct {
	CallStack              []TestEthBackCall
	Logs                   []types.Log
	BalanceEth             *big.Int
	BalancePSC             uint64
	Allowance              uint64
	BalancePTC             *big.Int
	BlockNumber            *big.Int
	Abi                    abi.ABI
	PscAddr                common.Address
	Tx                     *types.Transaction
	TxIsPending            bool
	OfferingAgent          common.Address
	OfferMinDeposit        uint64
	OfferCurrentSupply     uint16
	OfferMaxSupply         uint16
	OfferUpdateBlockNumber uint32
	OfferingIsActive       bool
	GasPrice               *big.Int
	EstimatedGas           uint64
}

var _ Backend = new(TestEthBackend)

// NewTestEthBackend creates new TestEthBackend instance.
func NewTestEthBackend(pscAddr common.Address) *TestEthBackend {
	b := &TestEthBackend{}
	b.PscAddr = pscAddr
	b.BlockNumber = big.NewInt(1)
	b.BalanceEth = big.NewInt(0)
	b.GasPrice = big.NewInt(20000000)
	return b
}

// PendingNonceAt returns the account nonce of the given account in the pending state.
// This is the nonce that should be used for the next transaction.
func (b *TestEthBackend) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return 123, nil
}

// LatestBlockNumber is mock to LatestBlockNumber.
func (b *TestEthBackend) LatestBlockNumber(ctx context.Context) (*big.Int, error) {
	block := b.BlockNumber
	b.BlockNumber = new(big.Int).Add(b.BlockNumber, big.NewInt(1))
	return block, nil
}

// SuggestGasPrice is mock to SuggestGasPrice.
func (b *TestEthBackend) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return b.GasPrice, nil
}

// SendTransaction is mock for send transaction.
func (b *TestEthBackend) SendTransaction(_ context.Context,
	_ *types.Transaction) error {
	return nil
}

// EstimateGas is mock to EstimateGas.
func (b *TestEthBackend) EstimateGas(
	ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	return b.EstimatedGas, nil
}

// CooperativeClose is mock to CooperativeClose.
func (b *TestEthBackend) CooperativeClose(opts *bind.TransactOpts,
	agentAddr common.Address, block uint32, offeringHash [32]byte,
	balance uint64, balanceMsgSig []byte, ClosingSig []byte) (*types.Transaction, error) {
	b.CallStack = append(b.CallStack, TestEthBackCall{
		method: "CooperativeClose",
		caller: opts.From,
		txOpts: opts,
		args: []interface{}{agentAddr, block, offeringHash, balance,
			balanceMsgSig, ClosingSig},
	})
	tx := types.NewTransaction(0, common.Address{}, big.NewInt(1), 1, big.NewInt(1), nil)
	return tx, nil
}

// RegisterServiceOffering is mock to RegisterServiceOffering.
func (b *TestEthBackend) RegisterServiceOffering(opts *bind.TransactOpts,
	offeringHash [32]byte, minDeposit uint64, maxSupply uint16,
	somcData uint8, somcType data.Base64String) (*types.Transaction, error) {
	b.CallStack = append(b.CallStack, TestEthBackCall{
		method: "RegisterServiceOffering",
		caller: opts.From,
		txOpts: opts,
		args: []interface{}{offeringHash, minDeposit, maxSupply,
			somcData, somcType},
	})
	b.OfferingAgent = opts.From
	b.OfferMinDeposit = minDeposit
	b.OfferMaxSupply = maxSupply
	b.OfferCurrentSupply = maxSupply
	b.OfferingIsActive = true

	nextBlock, _ := b.LatestBlockNumber(context.Background())
	b.BlockNumber = nextBlock

	tx := types.NewTransaction(0, common.Address{}, big.NewInt(1), 1, big.NewInt(1), nil)
	return tx, nil
}

// EthBalanceAt is mock to EthBalanceAt.
func (b *TestEthBackend) EthBalanceAt(_ context.Context,
	addr common.Address) (*big.Int, error) {
	b.CallStack = append(b.CallStack, TestEthBackCall{
		method: "EthBalanceAt",
		args:   []interface{}{addr},
	})
	return b.BalanceEth, nil
}

// PTCBalanceOf is mock to PTCBalanceOf.
func (b *TestEthBackend) PTCBalanceOf(opts *bind.CallOpts,
	addr common.Address) (*big.Int, error) {
	b.CallStack = append(b.CallStack, TestEthBackCall{
		method: "PTCBalanceOf",
		caller: opts.From,
		args:   []interface{}{addr},
	})
	return b.BalancePTC, nil
}

// PSCBalanceOf is mock to PSCBalanceOf.
func (b *TestEthBackend) PSCBalanceOf(opts *bind.CallOpts,
	addr common.Address) (uint64, error) {
	b.CallStack = append(b.CallStack, TestEthBackCall{
		method: "PSCBalanceOf",
		caller: opts.From,
		args:   []interface{}{addr},
	})
	return b.BalancePSC, nil
}

// PTCIncreaseApproval is mock to PTCIncreaseApproval.
func (b *TestEthBackend) PTCIncreaseApproval(opts *bind.TransactOpts,
	addr common.Address, amount *big.Int) (*types.Transaction, error) {
	b.CallStack = append(b.CallStack, TestEthBackCall{
		method: "PTCIncreaseApproval",
		caller: opts.From,
		txOpts: opts,
		args:   []interface{}{addr, amount},
	})
	tx := types.NewTransaction(0, common.Address{}, big.NewInt(1), 1, big.NewInt(1), nil)
	return tx, nil
}

func (b *TestEthBackend) PTCAllowance(opts *bind.CallOpts, owner, spender common.Address) (uint64, error) {
	b.CallStack = append(b.CallStack, TestEthBackCall{
		method: "PTCAllowance",
		caller: opts.From,
		args:   []interface{}{owner, spender},
	})
	return b.Allowance, nil
}

// PSCAddBalanceERC20 is mock to PSCAddBalanceERC20.
func (b *TestEthBackend) PSCAddBalanceERC20(opts *bind.TransactOpts,
	val uint64) (*types.Transaction, error) {
	b.CallStack = append(b.CallStack, TestEthBackCall{
		method: "PSCAddBalanceERC20",
		caller: opts.From,
		txOpts: opts,
		args:   []interface{}{val},
	})
	tx := types.NewTransaction(0, common.Address{}, big.NewInt(1), 1, big.NewInt(1), nil)
	return tx, nil
}

// PSCReturnBalanceERC20 is mock to PSCReturnBalanceERC20.
func (b *TestEthBackend) PSCReturnBalanceERC20(opts *bind.TransactOpts,
	val uint64) (*types.Transaction, error) {
	b.CallStack = append(b.CallStack, TestEthBackCall{
		method: "PSCReturnBalanceERC20",
		caller: opts.From,
		txOpts: opts,
		args:   []interface{}{val},
	})
	tx := types.NewTransaction(0, common.Address{}, big.NewInt(1), 1, big.NewInt(1), nil)
	return tx, nil
}

// PSCGetChannelInfo is mock to PSCGetChannelInfo.
func (b *TestEthBackend) PSCGetChannelInfo(opts *bind.CallOpts,
	client common.Address, agent common.Address,
	blockNumber uint32,
	hash [common.HashLength]byte) (uint64, uint32, uint64, error) {
	settleBlock, _ := b.LatestBlockNumber(context.Background())

	return 0, uint32(settleBlock.Uint64()), 0, nil
}

// SetTransaction mocks return value for GetTransactionByHash.
func (b *TestEthBackend) SetTransaction(t *testing.T,
	opts *bind.TransactOpts, input []byte) {
	rawTx := types.NewTransaction(1, b.PscAddr, nil, 0, nil, input)
	signedTx, err := opts.Signer(types.HomesteadSigner{},
		opts.From, rawTx)
	if err != nil {
		t.Fatal(err)
	}

	b.Tx = signedTx
}

// GetTransactionByHash is mock to GetTransactionByHash.
func (b *TestEthBackend) GetTransactionByHash(context.Context,
	common.Hash) (*types.Transaction, bool, error) {
	return b.Tx, b.TxIsPending, nil
}

// TestCalled tests the existence of a Ethereum call.
func (b *TestEthBackend) TestCalled(t *testing.T, method string,
	caller common.Address, gasLimit uint64, args ...interface{}) {
	if len(b.CallStack) == 0 {
		t.Fatalf("method %s not called. Callstack is empty", method)
	}
	for _, call := range b.CallStack {
		if caller == call.caller && method == call.method &&
			reflect.DeepEqual(args, call.args) &&
			(call.txOpts == nil || call.txOpts.GasLimit == gasLimit) {
			return
		}
	}
	t.Fatalf("no call of %s from %v with args: %v", method, caller, args)
}

// PSCGetOfferingInfo is mock to PSCGetOfferingInfo.
func (b *TestEthBackend) PSCGetOfferingInfo(opts *bind.CallOpts,
	hash [common.HashLength]byte) (agentAddr common.Address,
	minDeposit uint64, maxSupply uint16, currentSupply uint16,
	updateBlockNumber uint32, active bool, err error) {
	b.CallStack = append(b.CallStack, TestEthBackCall{
		method: "GetOfferingInfo",
		caller: opts.From,
		args:   []interface{}{hash},
	})
	return b.OfferingAgent, b.OfferMinDeposit, b.OfferMaxSupply,
		b.OfferCurrentSupply, b.OfferUpdateBlockNumber,
		b.OfferingIsActive, nil
}

// Test constants.
const (
	TestTXNonce    uint64 = 1
	TestTXGasLimit uint64 = 2
	TestTXGasPrice int64  = 1
)

// PSCCreateChannel is mock to PSCCreateChannel.
func (b *TestEthBackend) PSCCreateChannel(opts *bind.TransactOpts,
	agent common.Address, hash [common.HashLength]byte,
	deposit uint64) (*types.Transaction, error) {
	b.CallStack = append(b.CallStack, TestEthBackCall{
		method: "PSCCreateChannel",
		caller: opts.From,
		args:   []interface{}{agent, hash, deposit},
	})

	b.OfferCurrentSupply--

	tx := types.NewTransaction(
		TestTXNonce, agent, new(big.Int).SetUint64(deposit), TestTXGasLimit,
		big.NewInt(TestTXGasPrice), []byte{})

	return tx, nil
}

// PSCSettle is mock to PSCSettle.
func (b *TestEthBackend) PSCSettle(opts *bind.TransactOpts,
	agent common.Address, blockNumber uint32,
	hash [common.HashLength]byte) (*types.Transaction, error) {
	b.CallStack = append(b.CallStack, TestEthBackCall{
		method: "Settle",
		caller: opts.From,
		args:   []interface{}{agent, blockNumber, hash},
	})

	b.OfferCurrentSupply++

	tx := types.NewTransaction(
		TestTXNonce, agent, new(big.Int), TestTXGasLimit,
		big.NewInt(TestTXGasPrice), []byte{})

	return tx, nil
}

// PSCTopUpChannel is mock to PSCTopUpChannel.
func (b *TestEthBackend) PSCTopUpChannel(opts *bind.TransactOpts,
	agent common.Address, blockNumber uint32, hash [common.HashLength]byte,
	deposit uint64) (*types.Transaction, error) {
	b.CallStack = append(b.CallStack, TestEthBackCall{
		method: "TopUpChannel",
		caller: opts.From,
		args:   []interface{}{agent, blockNumber, hash, deposit},
	})

	tx := types.NewTransaction(
		TestTXNonce, agent, new(big.Int).SetUint64(deposit), TestTXGasLimit,
		opts.GasPrice, []byte{})

	return tx, nil
}

// PSCUncooperativeClose is mock to PSCUncooperativeClose.
func (b *TestEthBackend) PSCUncooperativeClose(opts *bind.TransactOpts,
	agent common.Address, blockNumber uint32, hash [common.HashLength]byte,
	balance uint64) (*types.Transaction, error) {
	b.CallStack = append(b.CallStack, TestEthBackCall{
		method: "UncooperativeClose",
		caller: opts.From,
		args:   []interface{}{agent, blockNumber, hash, balance},
	})

	tx := types.NewTransaction(
		TestTXNonce, agent, new(big.Int).SetUint64(balance), TestTXGasLimit,
		opts.GasPrice, []byte{})

	return tx, nil
}

// PSCRemoveServiceOffering is mock to PSCRemoveServiceOffering.
func (b *TestEthBackend) PSCRemoveServiceOffering(opts *bind.TransactOpts,
	offeringHash [32]byte) (*types.Transaction, error) {
	b.CallStack = append(b.CallStack, TestEthBackCall{
		txOpts: opts,
		method: "RemoveServiceOffering",
		caller: opts.From,
		args:   []interface{}{offeringHash},
	})

	b.OfferingIsActive = false

	tx := types.NewTransaction(
		TestTXNonce, b.PscAddr, nil, opts.GasLimit,
		opts.GasPrice, []byte{})

	return tx, nil
}

// PSCPopupServiceOffering is mock to PSCPopupServiceOffering.
func (b *TestEthBackend) PSCPopupServiceOffering(opts *bind.TransactOpts,
	offeringHash [32]byte, somcType uint8, somcData data.Base64String) (*types.Transaction, error) {
	b.CallStack = append(b.CallStack, TestEthBackCall{
		txOpts: opts,
		method: "PopupServiceOffering",
		caller: opts.From,
		args:   []interface{}{offeringHash, somcType, somcData},
	})

	nextBlock, _ := b.LatestBlockNumber(context.Background())
	b.OfferUpdateBlockNumber = uint32(nextBlock.Uint64())

	tx := types.NewTransaction(
		TestTXNonce, b.PscAddr, nil, opts.GasLimit,
		opts.GasPrice, []byte{})

	return tx, nil
}

// FilterLogs is mock to FilterLogs.
func (b *TestEthBackend) FilterLogs(ctx context.Context,
	q ethereum.FilterQuery) ([]types.Log, error) {
	return b.Logs, nil
}

// HeaderByNumber is mock to HeaderByNumber.
func (b *TestEthBackend) HeaderByNumber(ctx context.Context,
	number *big.Int) (*types.Header, error) {
	return &types.Header{
		Number: b.BlockNumber,
	}, nil
}

// PTCAddress is mock to PTCAddress.
func (b *TestEthBackend) PTCAddress() common.Address {
	return common.Address{}
}

// PSCAddress is mock to PSCAddress.
func (b *TestEthBackend) PSCAddress() common.Address {
	return b.PscAddr
}
