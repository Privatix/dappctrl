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
	method string
	caller common.Address
	args   []interface{}
}

type testEthBackend struct {
	callStack   []testEthBackCall
	balancePSC  *big.Int
	balancePTC  *big.Int
	abi         abi.ABI
	pscAddr     common.Address
	offerSupply uint16
	tx          *types.Transaction
}

func newTestEthBackend(pscAddr common.Address) *testEthBackend {
	b := &testEthBackend{}
	b.pscAddr = pscAddr
	return b
}

func (b *testEthBackend) CooperativeClose(opts *bind.TransactOpts,
	agentAddr common.Address, block uint32, offeringHash [32]byte,
	balance *big.Int, balanceMsgSig []byte, ClosingSig []byte) error {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "CooperativeClose",
		caller: opts.From,
		args: []interface{}{agentAddr, block, offeringHash, balance,
			balanceMsgSig, ClosingSig},
	})
	return nil
}

func (b *testEthBackend) RegisterServiceOffering(opts *bind.TransactOpts,
	offeringHash [32]byte, minDeposit *big.Int, maxSupply uint16) error {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "RegisterServiceOffering",
		caller: opts.From,
		args:   []interface{}{offeringHash, minDeposit, maxSupply},
	})
	return nil
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
	addr common.Address, amount *big.Int) error {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "PTCIncreaseApproval",
		caller: opts.From,
		args:   []interface{}{addr, amount},
	})
	return nil
}

func (b *testEthBackend) PSCAddBalanceERC20(opts *bind.TransactOpts,
	val *big.Int) error {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "PSCAddBalanceERC20",
		caller: opts.From,
		args:   []interface{}{val},
	})
	return nil
}

func (b *testEthBackend) PSCReturnBalanceERC20(opts *bind.TransactOpts,
	val *big.Int) error {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "PSCReturnBalanceERC20",
		caller: opts.From,
		args:   []interface{}{val},
	})
	return nil
}

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
	caller common.Address, args ...interface{}) {
	if len(b.callStack) == 0 {
		t.Fatalf("method %s not called. Callstack is empty", method)
	}
	for _, call := range b.callStack {
		if caller == call.caller && method == call.method &&
			reflect.DeepEqual(args, call.args) {
			return
		}
	}
	t.Logf("%+v\n", b.callStack)
	t.Fatalf("no call of %s from %v with args: %v", method, caller, args)
}

func (b *testEthBackend) PSCOfferingSupply(opts *bind.CallOpts,
	hash [common.HashLength]byte) (uint16, error) {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "PSCOfferingSupply",
		caller: opts.From,
		args:   []interface{}{hash},
	})
	return b.offerSupply, nil
}

func (b *testEthBackend) PSCCreateChannel(opts *bind.TransactOpts,
	agent common.Address, hash [common.HashLength]byte,
	deposit *big.Int) (*types.Transaction, error) {
	b.callStack = append(b.callStack, testEthBackCall{
		method: "PSCCreateChannel",
		caller: opts.From,
		args:   []interface{}{agent, hash, deposit},
	})
	return b.tx, nil
}
