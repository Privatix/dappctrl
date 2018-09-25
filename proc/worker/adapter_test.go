package worker

import (
	"context"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type testEthBackCall struct {
	txOpts *bind.TransactOpts
	method string
	caller common.Address
	args   []interface{}
}

type testEthBackend struct {
	callStack              []testEthBackCall
	balanceEth             *big.Int
	balancePSC             *big.Int
	balancePTC             *big.Int
	blockNumber            *big.Int
	abi                    abi.ABI
	pscAddr                common.Address
	tx                     *types.Transaction
	offeringAgent          common.Address
	offerMinDeposit        *big.Int
	offerCurrentSupply     uint16
	offerMaxSupply         uint16
	offerUpdateBlockNumber uint32
	offeringIsActive       bool
	challengePeriod        uint32
}

func newTestEthBackend(pscAddr common.Address) *testEthBackend {
	b := &testEthBackend{}
	b.pscAddr = pscAddr
	b.blockNumber = big.NewInt(1)
	return b
}

func (b *testEthBackend) LatestBlockNumber(ctx context.Context) (*big.Int, error) {
	block := b.blockNumber
	b.blockNumber = new(big.Int).Add(b.blockNumber, big.NewInt(1))
	return block, nil
}

func (b *testEthBackend) CooperativeClose(opts *bind.TransactOpts,
	agentAddr common.Address, block uint32, offeringHash [32]byte,
	balance *big.Int, balanceMsgSig []byte, ClosingSig []byte) (*types.Transaction, error) {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "CooperativeClose",
		caller: opts.From,
		txOpts: opts,
		args: []interface{}{agentAddr, block, offeringHash, balance,
			balanceMsgSig, ClosingSig},
	})
	tx := types.NewTransaction(0, common.Address{}, big.NewInt(1), 1, big.NewInt(1), nil)
	return tx, nil
}

func (b *testEthBackend) RegisterServiceOffering(opts *bind.TransactOpts,
	offeringHash [32]byte, minDeposit *big.Int, maxSupply uint16) (*types.Transaction, error) {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "RegisterServiceOffering",
		caller: opts.From,
		txOpts: opts,
		args:   []interface{}{offeringHash, minDeposit, maxSupply},
	})
	b.offeringAgent = opts.From
	b.offerMinDeposit = minDeposit
	b.offerMaxSupply = maxSupply
	b.offerCurrentSupply = maxSupply
	b.offeringIsActive = true

	nextBlock, _ := b.LatestBlockNumber(context.Background())
	b.blockNumber = nextBlock

	tx := types.NewTransaction(0, common.Address{}, big.NewInt(1), 1, big.NewInt(1), nil)
	return tx, nil
}

func (b *testEthBackend) EthBalanceAt(_ context.Context,
	addr common.Address) (*big.Int, error) {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "EthBalanceAt",
		args:   []interface{}{addr},
	})
	return b.balanceEth, nil
}

func (b *testEthBackend) PTCBalanceOf(opts *bind.CallOpts,
	addr common.Address) (*big.Int, error) {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "PTCBalanceOf",
		caller: opts.From,
		args:   []interface{}{addr},
	})
	return b.balancePTC, nil
}

func (b *testEthBackend) PSCBalanceOf(opts *bind.CallOpts,
	addr common.Address) (*big.Int, error) {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "PSCBalanceOf",
		caller: opts.From,
		args:   []interface{}{addr},
	})
	return b.balancePSC, nil
}

func (b *testEthBackend) PTCIncreaseApproval(opts *bind.TransactOpts,
	addr common.Address, amount *big.Int) (*types.Transaction, error) {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "PTCIncreaseApproval",
		caller: opts.From,
		txOpts: opts,
		args:   []interface{}{addr, amount},
	})
	tx := types.NewTransaction(0, common.Address{}, big.NewInt(1), 1, big.NewInt(1), nil)
	return tx, nil
}

func (b *testEthBackend) PSCAddBalanceERC20(opts *bind.TransactOpts,
	val *big.Int) (*types.Transaction, error) {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "PSCAddBalanceERC20",
		caller: opts.From,
		txOpts: opts,
		args:   []interface{}{val},
	})
	tx := types.NewTransaction(0, common.Address{}, big.NewInt(1), 1, big.NewInt(1), nil)
	return tx, nil
}

func (b *testEthBackend) PSCReturnBalanceERC20(opts *bind.TransactOpts,
	val *big.Int) (*types.Transaction, error) {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "PSCReturnBalanceERC20",
		caller: opts.From,
		txOpts: opts,
		args:   []interface{}{val},
	})
	tx := types.NewTransaction(0, common.Address{}, big.NewInt(1), 1, big.NewInt(1), nil)
	return tx, nil
}

func (b *testEthBackend) PSCGetChannelInfo(opts *bind.CallOpts,
	client common.Address, agent common.Address,
	blockNumber uint32,
	hash [common.HashLength]byte) ([common.HashLength]byte,
	*big.Int, uint32, *big.Int, error) {
	settleBlock, _ := b.LatestBlockNumber(context.Background())

	return [32]byte{}, nil, uint32(settleBlock.Uint64()), nil, nil
}

// setTransaction mocks return value for GetTransactionByHash.
func (b *testEthBackend) setTransaction(t *testing.T,
	opts *bind.TransactOpts, input []byte) {
	rawTx := types.NewTransaction(1, b.pscAddr, nil, 0, nil, input)
	signedTx, err := opts.Signer(types.HomesteadSigner{},
		opts.From, rawTx)
	if err != nil {
		t.Fatal(err)
	}

	b.tx = signedTx
}

func (b *testEthBackend) GetTransactionByHash(context.Context,
	common.Hash) (*types.Transaction, bool, error) {
	return b.tx, false, nil
}

func (b *testEthBackend) testCalled(t *testing.T, method string,
	caller common.Address, gasLimit uint64, args ...interface{}) {
	if len(b.callStack) == 0 {
		t.Fatalf("method %s not called. Callstack is empty", method)
	}
	for _, call := range b.callStack {
		if caller == call.caller && method == call.method &&
			reflect.DeepEqual(args, call.args) &&
			(call.txOpts == nil || call.txOpts.GasLimit == gasLimit) {
			return
		}
	}
	t.Logf("%+v\n", b.callStack)
	t.Fatalf("no call of %s from %v with args: %v", method, caller, args)
}

func (b *testEthBackend) PSCGetOfferingInfo(opts *bind.CallOpts,
	hash [common.HashLength]byte) (agentAddr common.Address,
	minDeposit *big.Int, maxSupply uint16, currentSupply uint16,
	updateBlockNumber uint32, active bool, err error) {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "GetOfferingInfo",
		caller: opts.From,
		args:   []interface{}{hash},
	})
	return b.offeringAgent, b.offerMinDeposit, b.offerMaxSupply,
		b.offerCurrentSupply, b.offerUpdateBlockNumber,
		b.offeringIsActive, nil
}

func (b *testEthBackend) PSCGetPopUpPeriod(
	opts *bind.CallOpts) (uint32, error) {
	return b.challengePeriod, nil
}

const (
	testTXNonce    uint64 = 1
	testTXGasLimit uint64 = 2
	testTXGasPrice int64  = 1
)

func (b *testEthBackend) PSCCreateChannel(opts *bind.TransactOpts,
	agent common.Address, hash [common.HashLength]byte,
	deposit *big.Int) (*types.Transaction, error) {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "PSCCreateChannel",
		caller: opts.From,
		args:   []interface{}{agent, hash, deposit},
	})

	b.offerCurrentSupply--

	tx := types.NewTransaction(
		testTXNonce, agent, deposit, testTXGasLimit,
		big.NewInt(testTXGasPrice), []byte{})

	return tx, nil
}

func (b *testEthBackend) PSCSettle(opts *bind.TransactOpts,
	agent common.Address, blockNumber uint32,
	hash [common.HashLength]byte) (*types.Transaction, error) {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "Settle",
		caller: opts.From,
		args:   []interface{}{agent, blockNumber, hash},
	})

	b.offerCurrentSupply++

	tx := types.NewTransaction(
		testTXNonce, agent, new(big.Int), testTXGasLimit,
		big.NewInt(testTXGasPrice), []byte{})

	return tx, nil
}

func (b *testEthBackend) PSCTopUpChannel(opts *bind.TransactOpts,
	agent common.Address, blockNumber uint32, hash [common.HashLength]byte,
	deposit *big.Int) (*types.Transaction, error) {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "TopUpChannel",
		caller: opts.From,
		args:   []interface{}{agent, blockNumber, hash, deposit},
	})

	tx := types.NewTransaction(
		testTXNonce, agent, deposit, testTXGasLimit,
		opts.GasPrice, []byte{})

	return tx, nil
}

func (b *testEthBackend) PSCUncooperativeClose(opts *bind.TransactOpts,
	agent common.Address, blockNumber uint32, hash [common.HashLength]byte,
	balance *big.Int) (*types.Transaction, error) {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "UncooperativeClose",
		caller: opts.From,
		args:   []interface{}{agent, blockNumber, hash, balance},
	})

	tx := types.NewTransaction(
		testTXNonce, agent, balance, testTXGasLimit,
		opts.GasPrice, []byte{})

	return tx, nil
}

func (b *testEthBackend) PSCRemoveServiceOffering(opts *bind.TransactOpts,
	offeringHash [32]byte) (*types.Transaction, error) {
	b.callStack = append(b.callStack, testEthBackCall{
		txOpts: opts,
		method: "RemoveServiceOffering",
		caller: opts.From,
		args:   []interface{}{offeringHash},
	})

	b.offeringIsActive = false

	tx := types.NewTransaction(
		testTXNonce, b.pscAddr, nil, opts.GasLimit,
		opts.GasPrice, []byte{})

	return tx, nil
}

func (b *testEthBackend) PSCPopupServiceOffering(opts *bind.TransactOpts,
	offeringHash [32]byte) (*types.Transaction, error) {
	b.callStack = append(b.callStack, testEthBackCall{
		txOpts: opts,
		method: "PopupServiceOffering",
		caller: opts.From,
		args:   []interface{}{offeringHash},
	})

	nextBlock, _ := b.LatestBlockNumber(context.Background())
	b.offerUpdateBlockNumber = uint32(nextBlock.Uint64())

	tx := types.NewTransaction(
		testTXNonce, b.pscAddr, nil, opts.GasLimit,
		opts.GasPrice, []byte{})

	return tx, nil
}
