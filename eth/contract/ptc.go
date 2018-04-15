// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// PrivatixTokenContractABI is the input ABI used to generate the binding from.
const PrivatixTokenContractABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"}]"

// PrivatixTokenContract is an auto generated Go binding around an Ethereum contract.
type PrivatixTokenContract struct {
	PrivatixTokenContractCaller     // Read-only binding to the contract
	PrivatixTokenContractTransactor // Write-only binding to the contract
	PrivatixTokenContractFilterer   // Log filterer for contract events
}

// PrivatixTokenContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type PrivatixTokenContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PrivatixTokenContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PrivatixTokenContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PrivatixTokenContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PrivatixTokenContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PrivatixTokenContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PrivatixTokenContractSession struct {
	Contract     *PrivatixTokenContract // Generic contract binding to set the session for
	CallOpts     bind.CallOpts          // Call options to use throughout this session
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// PrivatixTokenContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PrivatixTokenContractCallerSession struct {
	Contract *PrivatixTokenContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                // Call options to use throughout this session
}

// PrivatixTokenContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PrivatixTokenContractTransactorSession struct {
	Contract     *PrivatixTokenContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                // Transaction auth options to use throughout this session
}

// PrivatixTokenContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type PrivatixTokenContractRaw struct {
	Contract *PrivatixTokenContract // Generic contract binding to access the raw methods on
}

// PrivatixTokenContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PrivatixTokenContractCallerRaw struct {
	Contract *PrivatixTokenContractCaller // Generic read-only contract binding to access the raw methods on
}

// PrivatixTokenContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PrivatixTokenContractTransactorRaw struct {
	Contract *PrivatixTokenContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPrivatixTokenContract creates a new instance of PrivatixTokenContract, bound to a specific deployed contract.
func NewPrivatixTokenContract(address common.Address, backend bind.ContractBackend) (*PrivatixTokenContract, error) {
	contract, err := bindPrivatixTokenContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PrivatixTokenContract{PrivatixTokenContractCaller: PrivatixTokenContractCaller{contract: contract}, PrivatixTokenContractTransactor: PrivatixTokenContractTransactor{contract: contract}, PrivatixTokenContractFilterer: PrivatixTokenContractFilterer{contract: contract}}, nil
}

// NewPrivatixTokenContractCaller creates a new read-only instance of PrivatixTokenContract, bound to a specific deployed contract.
func NewPrivatixTokenContractCaller(address common.Address, caller bind.ContractCaller) (*PrivatixTokenContractCaller, error) {
	contract, err := bindPrivatixTokenContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PrivatixTokenContractCaller{contract: contract}, nil
}

// NewPrivatixTokenContractTransactor creates a new write-only instance of PrivatixTokenContract, bound to a specific deployed contract.
func NewPrivatixTokenContractTransactor(address common.Address, transactor bind.ContractTransactor) (*PrivatixTokenContractTransactor, error) {
	contract, err := bindPrivatixTokenContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PrivatixTokenContractTransactor{contract: contract}, nil
}

// NewPrivatixTokenContractFilterer creates a new log filterer instance of PrivatixTokenContract, bound to a specific deployed contract.
func NewPrivatixTokenContractFilterer(address common.Address, filterer bind.ContractFilterer) (*PrivatixTokenContractFilterer, error) {
	contract, err := bindPrivatixTokenContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PrivatixTokenContractFilterer{contract: contract}, nil
}

// bindPrivatixTokenContract binds a generic wrapper to an already deployed contract.
func bindPrivatixTokenContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(PrivatixTokenContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PrivatixTokenContract *PrivatixTokenContractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PrivatixTokenContract.Contract.PrivatixTokenContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PrivatixTokenContract *PrivatixTokenContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PrivatixTokenContract.Contract.PrivatixTokenContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PrivatixTokenContract *PrivatixTokenContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PrivatixTokenContract.Contract.PrivatixTokenContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PrivatixTokenContract *PrivatixTokenContractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PrivatixTokenContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PrivatixTokenContract *PrivatixTokenContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PrivatixTokenContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PrivatixTokenContract *PrivatixTokenContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PrivatixTokenContract.Contract.contract.Transact(opts, method, params...)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_PrivatixTokenContract *PrivatixTokenContractCaller) BalanceOf(opts *bind.CallOpts, _owner common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PrivatixTokenContract.contract.Call(opts, out, "balanceOf", _owner)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_PrivatixTokenContract *PrivatixTokenContractSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _PrivatixTokenContract.Contract.BalanceOf(&_PrivatixTokenContract.CallOpts, _owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_PrivatixTokenContract *PrivatixTokenContractCallerSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _PrivatixTokenContract.Contract.BalanceOf(&_PrivatixTokenContract.CallOpts, _owner)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_PrivatixTokenContract *PrivatixTokenContractCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PrivatixTokenContract.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_PrivatixTokenContract *PrivatixTokenContractSession) TotalSupply() (*big.Int, error) {
	return _PrivatixTokenContract.Contract.TotalSupply(&_PrivatixTokenContract.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_PrivatixTokenContract *PrivatixTokenContractCallerSession) TotalSupply() (*big.Int, error) {
	return _PrivatixTokenContract.Contract.TotalSupply(&_PrivatixTokenContract.CallOpts)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(bool)
func (_PrivatixTokenContract *PrivatixTokenContractTransactor) Transfer(opts *bind.TransactOpts, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _PrivatixTokenContract.contract.Transact(opts, "transfer", _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(bool)
func (_PrivatixTokenContract *PrivatixTokenContractSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _PrivatixTokenContract.Contract.Transfer(&_PrivatixTokenContract.TransactOpts, _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(bool)
func (_PrivatixTokenContract *PrivatixTokenContractTransactorSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _PrivatixTokenContract.Contract.Transfer(&_PrivatixTokenContract.TransactOpts, _to, _value)
}

// PrivatixTokenContractTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the PrivatixTokenContract contract.
type PrivatixTokenContractTransferIterator struct {
	Event *PrivatixTokenContractTransfer // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PrivatixTokenContractTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PrivatixTokenContractTransfer)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PrivatixTokenContractTransfer)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PrivatixTokenContractTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PrivatixTokenContractTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PrivatixTokenContractTransfer represents a Transfer event raised by the PrivatixTokenContract contract.
type PrivatixTokenContractTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(from indexed address, to indexed address, value uint256)
func (_PrivatixTokenContract *PrivatixTokenContractFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*PrivatixTokenContractTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _PrivatixTokenContract.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &PrivatixTokenContractTransferIterator{contract: _PrivatixTokenContract.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(from indexed address, to indexed address, value uint256)
func (_PrivatixTokenContract *PrivatixTokenContractFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *PrivatixTokenContractTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _PrivatixTokenContract.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PrivatixTokenContractTransfer)
				if err := _PrivatixTokenContract.contract.UnpackLog(event, "Transfer", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}
