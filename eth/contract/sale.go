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

// SaleABI is the input ABI used to generate the binding from.
const SaleABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"totalEthers\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"minimalEther\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"minFreePrixByTrans\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"endTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"wallet\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"maxFreePrixByTrans\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"grant\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"revoke\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"startTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"isOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"etherBalances\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"maximumTokens\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"weiPerToken\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"maybe_owner\",\"type\":\"address\"}],\"name\":\"checkOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"token\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_startTime\",\"type\":\"uint256\"},{\"name\":\"_wallet\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"TokenPurchase\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"AccessGrant\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"AccessRevoke\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"contributor\",\"type\":\"address\"}],\"name\":\"buyTokens\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"contributor\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"getFreeTokens\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"running\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// Sale is an auto generated Go binding around an Ethereum contract.
type Sale struct {
	SaleCaller     // Read-only binding to the contract
	SaleTransactor // Write-only binding to the contract
	SaleFilterer   // Log filterer for contract events
}

// SaleCaller is an auto generated read-only Go binding around an Ethereum contract.
type SaleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SaleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SaleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SaleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SaleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SaleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SaleSession struct {
	Contract     *Sale             // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SaleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SaleCallerSession struct {
	Contract *SaleCaller   // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// SaleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SaleTransactorSession struct {
	Contract     *SaleTransactor   // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SaleRaw is an auto generated low-level Go binding around an Ethereum contract.
type SaleRaw struct {
	Contract *Sale // Generic contract binding to access the raw methods on
}

// SaleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SaleCallerRaw struct {
	Contract *SaleCaller // Generic read-only contract binding to access the raw methods on
}

// SaleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SaleTransactorRaw struct {
	Contract *SaleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSale creates a new instance of Sale, bound to a specific deployed contract.
func NewSale(address common.Address, backend bind.ContractBackend) (*Sale, error) {
	contract, err := bindSale(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Sale{SaleCaller: SaleCaller{contract: contract}, SaleTransactor: SaleTransactor{contract: contract}, SaleFilterer: SaleFilterer{contract: contract}}, nil
}

// NewSaleCaller creates a new read-only instance of Sale, bound to a specific deployed contract.
func NewSaleCaller(address common.Address, caller bind.ContractCaller) (*SaleCaller, error) {
	contract, err := bindSale(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SaleCaller{contract: contract}, nil
}

// NewSaleTransactor creates a new write-only instance of Sale, bound to a specific deployed contract.
func NewSaleTransactor(address common.Address, transactor bind.ContractTransactor) (*SaleTransactor, error) {
	contract, err := bindSale(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SaleTransactor{contract: contract}, nil
}

// NewSaleFilterer creates a new log filterer instance of Sale, bound to a specific deployed contract.
func NewSaleFilterer(address common.Address, filterer bind.ContractFilterer) (*SaleFilterer, error) {
	contract, err := bindSale(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SaleFilterer{contract: contract}, nil
}

// bindSale binds a generic wrapper to an already deployed contract.
func bindSale(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SaleABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Sale *SaleRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Sale.Contract.SaleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Sale *SaleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Sale.Contract.SaleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Sale *SaleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Sale.Contract.SaleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Sale *SaleCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Sale.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Sale *SaleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Sale.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Sale *SaleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Sale.Contract.contract.Transact(opts, method, params...)
}

// CheckOwner is a free data retrieval call binding the contract method 0xe0e3671c.
//
// Solidity: function checkOwner(maybe_owner address) constant returns(bool)
func (_Sale *SaleCaller) CheckOwner(opts *bind.CallOpts, maybe_owner common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Sale.contract.Call(opts, out, "checkOwner", maybe_owner)
	return *ret0, err
}

// CheckOwner is a free data retrieval call binding the contract method 0xe0e3671c.
//
// Solidity: function checkOwner(maybe_owner address) constant returns(bool)
func (_Sale *SaleSession) CheckOwner(maybe_owner common.Address) (bool, error) {
	return _Sale.Contract.CheckOwner(&_Sale.CallOpts, maybe_owner)
}

// CheckOwner is a free data retrieval call binding the contract method 0xe0e3671c.
//
// Solidity: function checkOwner(maybe_owner address) constant returns(bool)
func (_Sale *SaleCallerSession) CheckOwner(maybe_owner common.Address) (bool, error) {
	return _Sale.Contract.CheckOwner(&_Sale.CallOpts, maybe_owner)
}

// EndTime is a free data retrieval call binding the contract method 0x3197cbb6.
//
// Solidity: function endTime() constant returns(uint256)
func (_Sale *SaleCaller) EndTime(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Sale.contract.Call(opts, out, "endTime")
	return *ret0, err
}

// EndTime is a free data retrieval call binding the contract method 0x3197cbb6.
//
// Solidity: function endTime() constant returns(uint256)
func (_Sale *SaleSession) EndTime() (*big.Int, error) {
	return _Sale.Contract.EndTime(&_Sale.CallOpts)
}

// EndTime is a free data retrieval call binding the contract method 0x3197cbb6.
//
// Solidity: function endTime() constant returns(uint256)
func (_Sale *SaleCallerSession) EndTime() (*big.Int, error) {
	return _Sale.Contract.EndTime(&_Sale.CallOpts)
}

// EtherBalances is a free data retrieval call binding the contract method 0x9653f8a1.
//
// Solidity: function etherBalances( address) constant returns(uint256)
func (_Sale *SaleCaller) EtherBalances(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Sale.contract.Call(opts, out, "etherBalances", arg0)
	return *ret0, err
}

// EtherBalances is a free data retrieval call binding the contract method 0x9653f8a1.
//
// Solidity: function etherBalances( address) constant returns(uint256)
func (_Sale *SaleSession) EtherBalances(arg0 common.Address) (*big.Int, error) {
	return _Sale.Contract.EtherBalances(&_Sale.CallOpts, arg0)
}

// EtherBalances is a free data retrieval call binding the contract method 0x9653f8a1.
//
// Solidity: function etherBalances( address) constant returns(uint256)
func (_Sale *SaleCallerSession) EtherBalances(arg0 common.Address) (*big.Int, error) {
	return _Sale.Contract.EtherBalances(&_Sale.CallOpts, arg0)
}

// IsOwner is a free data retrieval call binding the contract method 0x8f32d59b.
//
// Solidity: function isOwner() constant returns(bool)
func (_Sale *SaleCaller) IsOwner(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Sale.contract.Call(opts, out, "isOwner")
	return *ret0, err
}

// IsOwner is a free data retrieval call binding the contract method 0x8f32d59b.
//
// Solidity: function isOwner() constant returns(bool)
func (_Sale *SaleSession) IsOwner() (bool, error) {
	return _Sale.Contract.IsOwner(&_Sale.CallOpts)
}

// IsOwner is a free data retrieval call binding the contract method 0x8f32d59b.
//
// Solidity: function isOwner() constant returns(bool)
func (_Sale *SaleCallerSession) IsOwner() (bool, error) {
	return _Sale.Contract.IsOwner(&_Sale.CallOpts)
}

// MaxFreePrixByTrans is a free data retrieval call binding the contract method 0x62c5fa4b.
//
// Solidity: function maxFreePrixByTrans() constant returns(uint256)
func (_Sale *SaleCaller) MaxFreePrixByTrans(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Sale.contract.Call(opts, out, "maxFreePrixByTrans")
	return *ret0, err
}

// MaxFreePrixByTrans is a free data retrieval call binding the contract method 0x62c5fa4b.
//
// Solidity: function maxFreePrixByTrans() constant returns(uint256)
func (_Sale *SaleSession) MaxFreePrixByTrans() (*big.Int, error) {
	return _Sale.Contract.MaxFreePrixByTrans(&_Sale.CallOpts)
}

// MaxFreePrixByTrans is a free data retrieval call binding the contract method 0x62c5fa4b.
//
// Solidity: function maxFreePrixByTrans() constant returns(uint256)
func (_Sale *SaleCallerSession) MaxFreePrixByTrans() (*big.Int, error) {
	return _Sale.Contract.MaxFreePrixByTrans(&_Sale.CallOpts)
}

// MaximumTokens is a free data retrieval call binding the contract method 0xc3a9bd8b.
//
// Solidity: function maximumTokens() constant returns(uint256)
func (_Sale *SaleCaller) MaximumTokens(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Sale.contract.Call(opts, out, "maximumTokens")
	return *ret0, err
}

// MaximumTokens is a free data retrieval call binding the contract method 0xc3a9bd8b.
//
// Solidity: function maximumTokens() constant returns(uint256)
func (_Sale *SaleSession) MaximumTokens() (*big.Int, error) {
	return _Sale.Contract.MaximumTokens(&_Sale.CallOpts)
}

// MaximumTokens is a free data retrieval call binding the contract method 0xc3a9bd8b.
//
// Solidity: function maximumTokens() constant returns(uint256)
func (_Sale *SaleCallerSession) MaximumTokens() (*big.Int, error) {
	return _Sale.Contract.MaximumTokens(&_Sale.CallOpts)
}

// MinFreePrixByTrans is a free data retrieval call binding the contract method 0x22b96b42.
//
// Solidity: function minFreePrixByTrans() constant returns(uint256)
func (_Sale *SaleCaller) MinFreePrixByTrans(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Sale.contract.Call(opts, out, "minFreePrixByTrans")
	return *ret0, err
}

// MinFreePrixByTrans is a free data retrieval call binding the contract method 0x22b96b42.
//
// Solidity: function minFreePrixByTrans() constant returns(uint256)
func (_Sale *SaleSession) MinFreePrixByTrans() (*big.Int, error) {
	return _Sale.Contract.MinFreePrixByTrans(&_Sale.CallOpts)
}

// MinFreePrixByTrans is a free data retrieval call binding the contract method 0x22b96b42.
//
// Solidity: function minFreePrixByTrans() constant returns(uint256)
func (_Sale *SaleCallerSession) MinFreePrixByTrans() (*big.Int, error) {
	return _Sale.Contract.MinFreePrixByTrans(&_Sale.CallOpts)
}

// MinimalEther is a free data retrieval call binding the contract method 0x217ad35a.
//
// Solidity: function minimalEther() constant returns(uint256)
func (_Sale *SaleCaller) MinimalEther(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Sale.contract.Call(opts, out, "minimalEther")
	return *ret0, err
}

// MinimalEther is a free data retrieval call binding the contract method 0x217ad35a.
//
// Solidity: function minimalEther() constant returns(uint256)
func (_Sale *SaleSession) MinimalEther() (*big.Int, error) {
	return _Sale.Contract.MinimalEther(&_Sale.CallOpts)
}

// MinimalEther is a free data retrieval call binding the contract method 0x217ad35a.
//
// Solidity: function minimalEther() constant returns(uint256)
func (_Sale *SaleCallerSession) MinimalEther() (*big.Int, error) {
	return _Sale.Contract.MinimalEther(&_Sale.CallOpts)
}

// Running is a free data retrieval call binding the contract method 0xd85bd526.
//
// Solidity: function running() constant returns(bool)
func (_Sale *SaleCaller) Running(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Sale.contract.Call(opts, out, "running")
	return *ret0, err
}

// Running is a free data retrieval call binding the contract method 0xd85bd526.
//
// Solidity: function running() constant returns(bool)
func (_Sale *SaleSession) Running() (bool, error) {
	return _Sale.Contract.Running(&_Sale.CallOpts)
}

// Running is a free data retrieval call binding the contract method 0xd85bd526.
//
// Solidity: function running() constant returns(bool)
func (_Sale *SaleCallerSession) Running() (bool, error) {
	return _Sale.Contract.Running(&_Sale.CallOpts)
}

// StartTime is a free data retrieval call binding the contract method 0x78e97925.
//
// Solidity: function startTime() constant returns(uint256)
func (_Sale *SaleCaller) StartTime(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Sale.contract.Call(opts, out, "startTime")
	return *ret0, err
}

// StartTime is a free data retrieval call binding the contract method 0x78e97925.
//
// Solidity: function startTime() constant returns(uint256)
func (_Sale *SaleSession) StartTime() (*big.Int, error) {
	return _Sale.Contract.StartTime(&_Sale.CallOpts)
}

// StartTime is a free data retrieval call binding the contract method 0x78e97925.
//
// Solidity: function startTime() constant returns(uint256)
func (_Sale *SaleCallerSession) StartTime() (*big.Int, error) {
	return _Sale.Contract.StartTime(&_Sale.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() constant returns(address)
func (_Sale *SaleCaller) Token(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Sale.contract.Call(opts, out, "token")
	return *ret0, err
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() constant returns(address)
func (_Sale *SaleSession) Token() (common.Address, error) {
	return _Sale.Contract.Token(&_Sale.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() constant returns(address)
func (_Sale *SaleCallerSession) Token() (common.Address, error) {
	return _Sale.Contract.Token(&_Sale.CallOpts)
}

// TotalEthers is a free data retrieval call binding the contract method 0x0a4625af.
//
// Solidity: function totalEthers() constant returns(uint256)
func (_Sale *SaleCaller) TotalEthers(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Sale.contract.Call(opts, out, "totalEthers")
	return *ret0, err
}

// TotalEthers is a free data retrieval call binding the contract method 0x0a4625af.
//
// Solidity: function totalEthers() constant returns(uint256)
func (_Sale *SaleSession) TotalEthers() (*big.Int, error) {
	return _Sale.Contract.TotalEthers(&_Sale.CallOpts)
}

// TotalEthers is a free data retrieval call binding the contract method 0x0a4625af.
//
// Solidity: function totalEthers() constant returns(uint256)
func (_Sale *SaleCallerSession) TotalEthers() (*big.Int, error) {
	return _Sale.Contract.TotalEthers(&_Sale.CallOpts)
}

// Wallet is a free data retrieval call binding the contract method 0x521eb273.
//
// Solidity: function wallet() constant returns(address)
func (_Sale *SaleCaller) Wallet(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Sale.contract.Call(opts, out, "wallet")
	return *ret0, err
}

// Wallet is a free data retrieval call binding the contract method 0x521eb273.
//
// Solidity: function wallet() constant returns(address)
func (_Sale *SaleSession) Wallet() (common.Address, error) {
	return _Sale.Contract.Wallet(&_Sale.CallOpts)
}

// Wallet is a free data retrieval call binding the contract method 0x521eb273.
//
// Solidity: function wallet() constant returns(address)
func (_Sale *SaleCallerSession) Wallet() (common.Address, error) {
	return _Sale.Contract.Wallet(&_Sale.CallOpts)
}

// WeiPerToken is a free data retrieval call binding the contract method 0xdab8263a.
//
// Solidity: function weiPerToken() constant returns(uint256)
func (_Sale *SaleCaller) WeiPerToken(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Sale.contract.Call(opts, out, "weiPerToken")
	return *ret0, err
}

// WeiPerToken is a free data retrieval call binding the contract method 0xdab8263a.
//
// Solidity: function weiPerToken() constant returns(uint256)
func (_Sale *SaleSession) WeiPerToken() (*big.Int, error) {
	return _Sale.Contract.WeiPerToken(&_Sale.CallOpts)
}

// WeiPerToken is a free data retrieval call binding the contract method 0xdab8263a.
//
// Solidity: function weiPerToken() constant returns(uint256)
func (_Sale *SaleCallerSession) WeiPerToken() (*big.Int, error) {
	return _Sale.Contract.WeiPerToken(&_Sale.CallOpts)
}

// BuyTokens is a paid mutator transaction binding the contract method 0xec8ac4d8.
//
// Solidity: function buyTokens(contributor address) returns()
func (_Sale *SaleTransactor) BuyTokens(opts *bind.TransactOpts, contributor common.Address) (*types.Transaction, error) {
	return _Sale.contract.Transact(opts, "buyTokens", contributor)
}

// BuyTokens is a paid mutator transaction binding the contract method 0xec8ac4d8.
//
// Solidity: function buyTokens(contributor address) returns()
func (_Sale *SaleSession) BuyTokens(contributor common.Address) (*types.Transaction, error) {
	return _Sale.Contract.BuyTokens(&_Sale.TransactOpts, contributor)
}

// BuyTokens is a paid mutator transaction binding the contract method 0xec8ac4d8.
//
// Solidity: function buyTokens(contributor address) returns()
func (_Sale *SaleTransactorSession) BuyTokens(contributor common.Address) (*types.Transaction, error) {
	return _Sale.Contract.BuyTokens(&_Sale.TransactOpts, contributor)
}

// GetFreeTokens is a paid mutator transaction binding the contract method 0xccddff47.
//
// Solidity: function getFreeTokens(contributor address, amount uint256) returns()
func (_Sale *SaleTransactor) GetFreeTokens(opts *bind.TransactOpts, contributor common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Sale.contract.Transact(opts, "getFreeTokens", contributor, amount)
}

// GetFreeTokens is a paid mutator transaction binding the contract method 0xccddff47.
//
// Solidity: function getFreeTokens(contributor address, amount uint256) returns()
func (_Sale *SaleSession) GetFreeTokens(contributor common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Sale.Contract.GetFreeTokens(&_Sale.TransactOpts, contributor, amount)
}

// GetFreeTokens is a paid mutator transaction binding the contract method 0xccddff47.
//
// Solidity: function getFreeTokens(contributor address, amount uint256) returns()
func (_Sale *SaleTransactorSession) GetFreeTokens(contributor common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Sale.Contract.GetFreeTokens(&_Sale.TransactOpts, contributor, amount)
}

// Grant is a paid mutator transaction binding the contract method 0x70284d19.
//
// Solidity: function grant(_owner address) returns()
func (_Sale *SaleTransactor) Grant(opts *bind.TransactOpts, _owner common.Address) (*types.Transaction, error) {
	return _Sale.contract.Transact(opts, "grant", _owner)
}

// Grant is a paid mutator transaction binding the contract method 0x70284d19.
//
// Solidity: function grant(_owner address) returns()
func (_Sale *SaleSession) Grant(_owner common.Address) (*types.Transaction, error) {
	return _Sale.Contract.Grant(&_Sale.TransactOpts, _owner)
}

// Grant is a paid mutator transaction binding the contract method 0x70284d19.
//
// Solidity: function grant(_owner address) returns()
func (_Sale *SaleTransactorSession) Grant(_owner common.Address) (*types.Transaction, error) {
	return _Sale.Contract.Grant(&_Sale.TransactOpts, _owner)
}

// Revoke is a paid mutator transaction binding the contract method 0x74a8f103.
//
// Solidity: function revoke(_owner address) returns()
func (_Sale *SaleTransactor) Revoke(opts *bind.TransactOpts, _owner common.Address) (*types.Transaction, error) {
	return _Sale.contract.Transact(opts, "revoke", _owner)
}

// Revoke is a paid mutator transaction binding the contract method 0x74a8f103.
//
// Solidity: function revoke(_owner address) returns()
func (_Sale *SaleSession) Revoke(_owner common.Address) (*types.Transaction, error) {
	return _Sale.Contract.Revoke(&_Sale.TransactOpts, _owner)
}

// Revoke is a paid mutator transaction binding the contract method 0x74a8f103.
//
// Solidity: function revoke(_owner address) returns()
func (_Sale *SaleTransactorSession) Revoke(_owner common.Address) (*types.Transaction, error) {
	return _Sale.Contract.Revoke(&_Sale.TransactOpts, _owner)
}

// SaleAccessGrantIterator is returned from FilterAccessGrant and is used to iterate over the raw logs and unpacked data for AccessGrant events raised by the Sale contract.
type SaleAccessGrantIterator struct {
	Event *SaleAccessGrant // Event containing the contract specifics and raw log

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
func (it *SaleAccessGrantIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SaleAccessGrant)
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
		it.Event = new(SaleAccessGrant)
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
func (it *SaleAccessGrantIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SaleAccessGrantIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SaleAccessGrant represents a AccessGrant event raised by the Sale contract.
type SaleAccessGrant struct {
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterAccessGrant is a free log retrieval operation binding the contract event 0x1350a997c6c86bcc51dd7e51f7ef618d620e6a85d8fdabb82a980c149ad88d47.
//
// Solidity: event AccessGrant(owner indexed address)
func (_Sale *SaleFilterer) FilterAccessGrant(opts *bind.FilterOpts, owner []common.Address) (*SaleAccessGrantIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Sale.contract.FilterLogs(opts, "AccessGrant", ownerRule)
	if err != nil {
		return nil, err
	}
	return &SaleAccessGrantIterator{contract: _Sale.contract, event: "AccessGrant", logs: logs, sub: sub}, nil
}

// WatchAccessGrant is a free log subscription operation binding the contract event 0x1350a997c6c86bcc51dd7e51f7ef618d620e6a85d8fdabb82a980c149ad88d47.
//
// Solidity: event AccessGrant(owner indexed address)
func (_Sale *SaleFilterer) WatchAccessGrant(opts *bind.WatchOpts, sink chan<- *SaleAccessGrant, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Sale.contract.WatchLogs(opts, "AccessGrant", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SaleAccessGrant)
				if err := _Sale.contract.UnpackLog(event, "AccessGrant", log); err != nil {
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

// SaleAccessRevokeIterator is returned from FilterAccessRevoke and is used to iterate over the raw logs and unpacked data for AccessRevoke events raised by the Sale contract.
type SaleAccessRevokeIterator struct {
	Event *SaleAccessRevoke // Event containing the contract specifics and raw log

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
func (it *SaleAccessRevokeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SaleAccessRevoke)
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
		it.Event = new(SaleAccessRevoke)
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
func (it *SaleAccessRevokeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SaleAccessRevokeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SaleAccessRevoke represents a AccessRevoke event raised by the Sale contract.
type SaleAccessRevoke struct {
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterAccessRevoke is a free log retrieval operation binding the contract event 0x1d1eff42eefbeecfca7e39f8adb5d7f19a7ebbb4c3e82c51f2500d7d76ab2468.
//
// Solidity: event AccessRevoke(owner indexed address)
func (_Sale *SaleFilterer) FilterAccessRevoke(opts *bind.FilterOpts, owner []common.Address) (*SaleAccessRevokeIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Sale.contract.FilterLogs(opts, "AccessRevoke", ownerRule)
	if err != nil {
		return nil, err
	}
	return &SaleAccessRevokeIterator{contract: _Sale.contract, event: "AccessRevoke", logs: logs, sub: sub}, nil
}

// WatchAccessRevoke is a free log subscription operation binding the contract event 0x1d1eff42eefbeecfca7e39f8adb5d7f19a7ebbb4c3e82c51f2500d7d76ab2468.
//
// Solidity: event AccessRevoke(owner indexed address)
func (_Sale *SaleFilterer) WatchAccessRevoke(opts *bind.WatchOpts, sink chan<- *SaleAccessRevoke, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Sale.contract.WatchLogs(opts, "AccessRevoke", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SaleAccessRevoke)
				if err := _Sale.contract.UnpackLog(event, "AccessRevoke", log); err != nil {
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

// SaleTokenPurchaseIterator is returned from FilterTokenPurchase and is used to iterate over the raw logs and unpacked data for TokenPurchase events raised by the Sale contract.
type SaleTokenPurchaseIterator struct {
	Event *SaleTokenPurchase // Event containing the contract specifics and raw log

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
func (it *SaleTokenPurchaseIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SaleTokenPurchase)
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
		it.Event = new(SaleTokenPurchase)
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
func (it *SaleTokenPurchaseIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SaleTokenPurchaseIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SaleTokenPurchase represents a TokenPurchase event raised by the Sale contract.
type SaleTokenPurchase struct {
	Beneficiary common.Address
	Value       *big.Int
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterTokenPurchase is a free log retrieval operation binding the contract event 0xcd60aa75dea3072fbc07ae6d7d856b5dc5f4eee88854f5b4abf7b680ef8bc50f.
//
// Solidity: event TokenPurchase(beneficiary indexed address, value uint256, amount uint256)
func (_Sale *SaleFilterer) FilterTokenPurchase(opts *bind.FilterOpts, beneficiary []common.Address) (*SaleTokenPurchaseIterator, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _Sale.contract.FilterLogs(opts, "TokenPurchase", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return &SaleTokenPurchaseIterator{contract: _Sale.contract, event: "TokenPurchase", logs: logs, sub: sub}, nil
}

// WatchTokenPurchase is a free log subscription operation binding the contract event 0xcd60aa75dea3072fbc07ae6d7d856b5dc5f4eee88854f5b4abf7b680ef8bc50f.
//
// Solidity: event TokenPurchase(beneficiary indexed address, value uint256, amount uint256)
func (_Sale *SaleFilterer) WatchTokenPurchase(opts *bind.WatchOpts, sink chan<- *SaleTokenPurchase, beneficiary []common.Address) (event.Subscription, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _Sale.contract.WatchLogs(opts, "TokenPurchase", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SaleTokenPurchase)
				if err := _Sale.contract.UnpackLog(event, "TokenPurchase", log); err != nil {
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
