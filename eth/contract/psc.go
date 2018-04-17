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

// PrivatixServiceContractABI is the input ABI used to generate the binding from.
const PrivatixServiceContractABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"challenge_period\",\"outputs\":[{\"name\":\"\",\"type\":\"uint32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"internal_balances\",\"outputs\":[{\"name\":\"\",\"type\":\"uint192\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"channel_deposit_bugbounty_limit\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"network_fee\",\"outputs\":[{\"name\":\"\",\"type\":\"uint32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"closing_requests\",\"outputs\":[{\"name\":\"closing_balance\",\"type\":\"uint192\"},{\"name\":\"settle_block_number\",\"type\":\"uint32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"channels\",\"outputs\":[{\"name\":\"deposit\",\"type\":\"uint192\"},{\"name\":\"open_block_number\",\"type\":\"uint32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"meta_version\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"token\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_token_address\",\"type\":\"address\"},{\"name\":\"_network_fee_address\",\"type\":\"address\"},{\"name\":\"_challenge_period\",\"type\":\"uint32\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_agent\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_client\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"_deposit\",\"type\":\"uint192\"},{\"indexed\":false,\"name\":\"_authentication_hash\",\"type\":\"bytes32\"}],\"name\":\"LogChannelCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_agent\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_client\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"_open_block_number\",\"type\":\"uint32\"},{\"indexed\":false,\"name\":\"_added_deposit\",\"type\":\"uint192\"}],\"name\":\"LogChannelToppedUp\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_agent\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_client\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"_open_block_number\",\"type\":\"uint32\"},{\"indexed\":false,\"name\":\"_balance\",\"type\":\"uint192\"}],\"name\":\"LogChannelCloseRequested\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_agent\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"_min_deposit\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"_current_supply\",\"type\":\"uint16\"}],\"name\":\"LogOfferingCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_agent\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_offering_hash\",\"type\":\"bytes32\"}],\"name\":\"LogOfferingDeleted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_agent\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_client\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"_open_block_number\",\"type\":\"uint32\"},{\"indexed\":false,\"name\":\"_endpoint_hash\",\"type\":\"bytes32\"}],\"name\":\"LogOfferingEndpoint\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_agent\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"_current_supply\",\"type\":\"uint16\"}],\"name\":\"LogOfferingSupplyChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_agent\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_offering_hash\",\"type\":\"bytes32\"}],\"name\":\"LogOfferingPopedUp\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_agent\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_client\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"_open_block_number\",\"type\":\"uint32\"},{\"indexed\":false,\"name\":\"_balance\",\"type\":\"uint192\"}],\"name\":\"LogCooperativeChannelClose\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_agent\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_client\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"_open_block_number\",\"type\":\"uint32\"},{\"indexed\":false,\"name\":\"_balance\",\"type\":\"uint192\"}],\"name\":\"LogUnCooperativeChannelClose\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_value\",\"type\":\"uint192\"}],\"name\":\"addBalanceERC20\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_value\",\"type\":\"uint192\"}],\"name\":\"returnBalanceERC20\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_address\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_network_fee_address\",\"type\":\"address\"}],\"name\":\"setNetworkFeeAddress\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_network_fee\",\"type\":\"uint32\"}],\"name\":\"setNetworkFee\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_agent_address\",\"type\":\"address\"},{\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"name\":\"_deposit\",\"type\":\"uint192\"},{\"name\":\"_authentication_hash\",\"type\":\"bytes32\"}],\"name\":\"createChannel\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_agent_address\",\"type\":\"address\"},{\"name\":\"_open_block_number\",\"type\":\"uint32\"},{\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"name\":\"_added_deposit\",\"type\":\"uint192\"}],\"name\":\"topUpChannel\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_agent_address\",\"type\":\"address\"},{\"name\":\"_open_block_number\",\"type\":\"uint32\"},{\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"name\":\"_balance\",\"type\":\"uint192\"},{\"name\":\"_balance_msg_sig\",\"type\":\"bytes\"},{\"name\":\"_closing_sig\",\"type\":\"bytes\"}],\"name\":\"cooperativeClose\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_agent_address\",\"type\":\"address\"},{\"name\":\"_open_block_number\",\"type\":\"uint32\"},{\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"name\":\"_balance\",\"type\":\"uint192\"}],\"name\":\"uncooperativeClose\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_agent_address\",\"type\":\"address\"},{\"name\":\"_open_block_number\",\"type\":\"uint32\"},{\"name\":\"_offering_hash\",\"type\":\"bytes32\"}],\"name\":\"settle\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_client_address\",\"type\":\"address\"},{\"name\":\"_agent_address\",\"type\":\"address\"},{\"name\":\"_open_block_number\",\"type\":\"uint32\"},{\"name\":\"_offering_hash\",\"type\":\"bytes32\"}],\"name\":\"getChannelInfo\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"},{\"name\":\"\",\"type\":\"uint192\"},{\"name\":\"\",\"type\":\"uint32\"},{\"name\":\"\",\"type\":\"uint192\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_client_address\",\"type\":\"address\"},{\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"name\":\"_open_block_number\",\"type\":\"uint32\"},{\"name\":\"_endpoint_hash\",\"type\":\"bytes32\"}],\"name\":\"publishServiceOfferingEndpoint\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"name\":\"_min_deposit\",\"type\":\"uint192\"},{\"name\":\"_max_supply\",\"type\":\"uint16\"}],\"name\":\"registerServiceOffering\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_offering_hash\",\"type\":\"bytes32\"}],\"name\":\"removeServiceOffering\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_offering_hash\",\"type\":\"bytes32\"}],\"name\":\"popupServiceOffering\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_address\",\"type\":\"address\"},{\"name\":\"_open_block_number\",\"type\":\"uint32\"},{\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"name\":\"_balance\",\"type\":\"uint192\"},{\"name\":\"_msg_sig\",\"type\":\"bytes\"},{\"name\":\"_type\",\"type\":\"bool\"}],\"name\":\"extractSignature\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_client_address\",\"type\":\"address\"},{\"name\":\"_agent_address\",\"type\":\"address\"},{\"name\":\"_open_block_number\",\"type\":\"uint32\"},{\"name\":\"_offering_hash\",\"type\":\"bytes32\"}],\"name\":\"getKey\",\"outputs\":[{\"name\":\"data\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_agent_address\",\"type\":\"address\"},{\"name\":\"_client_address\",\"type\":\"address\"},{\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"name\":\"_deposit\",\"type\":\"uint192\"},{\"name\":\"_authentication_hash\",\"type\":\"bytes32\"}],\"name\":\"throwEventLogChannelCreated\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_agent_address\",\"type\":\"address\"},{\"name\":\"_client_address\",\"type\":\"address\"},{\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"name\":\"_open_block_number\",\"type\":\"uint32\"},{\"name\":\"_added_deposit\",\"type\":\"uint192\"}],\"name\":\"throwEventLogChannelToppedUp\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_agent_address\",\"type\":\"address\"},{\"name\":\"_client_address\",\"type\":\"address\"},{\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"name\":\"_open_block_number\",\"type\":\"uint32\"},{\"name\":\"_balance\",\"type\":\"uint192\"}],\"name\":\"throwEventLogChannelCloseRequested\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_agent_address\",\"type\":\"address\"},{\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"name\":\"_min_deposit\",\"type\":\"uint256\"},{\"name\":\"_current_supply\",\"type\":\"uint16\"}],\"name\":\"throwEventLogOfferingCreated\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_agent_address\",\"type\":\"address\"},{\"name\":\"_offering_hash\",\"type\":\"bytes32\"}],\"name\":\"throwEventLogOfferingDeleted\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_agent_address\",\"type\":\"address\"},{\"name\":\"_client_address\",\"type\":\"address\"},{\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"name\":\"_open_block_number\",\"type\":\"uint32\"},{\"name\":\"_endpoint_hash\",\"type\":\"bytes32\"}],\"name\":\"throwEventLogOfferingEndpoint\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_agent_address\",\"type\":\"address\"},{\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"name\":\"_current_supply\",\"type\":\"uint16\"}],\"name\":\"throwEventLogOfferingSupplyChanged\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_agent_address\",\"type\":\"address\"},{\"name\":\"_offering_hash\",\"type\":\"bytes32\"}],\"name\":\"throwEventLogOfferingPopedUp\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_agent_address\",\"type\":\"address\"},{\"name\":\"_client_address\",\"type\":\"address\"},{\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"name\":\"_open_block_number\",\"type\":\"uint32\"},{\"name\":\"_balance\",\"type\":\"uint192\"}],\"name\":\"throwEventLogCooperativeChannelClose\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_agent_address\",\"type\":\"address\"},{\"name\":\"_client_address\",\"type\":\"address\"},{\"name\":\"_offering_hash\",\"type\":\"bytes32\"},{\"name\":\"_open_block_number\",\"type\":\"uint32\"},{\"name\":\"_balance\",\"type\":\"uint192\"}],\"name\":\"throwEventLogUnCooperativeChannelClose\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// PrivatixServiceContract is an auto generated Go binding around an Ethereum contract.
type PrivatixServiceContract struct {
	PrivatixServiceContractCaller     // Read-only binding to the contract
	PrivatixServiceContractTransactor // Write-only binding to the contract
	PrivatixServiceContractFilterer   // Log filterer for contract events
}

// PrivatixServiceContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type PrivatixServiceContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PrivatixServiceContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PrivatixServiceContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PrivatixServiceContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PrivatixServiceContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PrivatixServiceContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PrivatixServiceContractSession struct {
	Contract     *PrivatixServiceContract // Generic contract binding to set the session for
	CallOpts     bind.CallOpts            // Call options to use throughout this session
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// PrivatixServiceContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PrivatixServiceContractCallerSession struct {
	Contract *PrivatixServiceContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                  // Call options to use throughout this session
}

// PrivatixServiceContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PrivatixServiceContractTransactorSession struct {
	Contract     *PrivatixServiceContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                  // Transaction auth options to use throughout this session
}

// PrivatixServiceContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type PrivatixServiceContractRaw struct {
	Contract *PrivatixServiceContract // Generic contract binding to access the raw methods on
}

// PrivatixServiceContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PrivatixServiceContractCallerRaw struct {
	Contract *PrivatixServiceContractCaller // Generic read-only contract binding to access the raw methods on
}

// PrivatixServiceContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PrivatixServiceContractTransactorRaw struct {
	Contract *PrivatixServiceContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPrivatixServiceContract creates a new instance of PrivatixServiceContract, bound to a specific deployed contract.
func NewPrivatixServiceContract(address common.Address, backend bind.ContractBackend) (*PrivatixServiceContract, error) {
	contract, err := bindPrivatixServiceContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PrivatixServiceContract{PrivatixServiceContractCaller: PrivatixServiceContractCaller{contract: contract}, PrivatixServiceContractTransactor: PrivatixServiceContractTransactor{contract: contract}, PrivatixServiceContractFilterer: PrivatixServiceContractFilterer{contract: contract}}, nil
}

// NewPrivatixServiceContractCaller creates a new read-only instance of PrivatixServiceContract, bound to a specific deployed contract.
func NewPrivatixServiceContractCaller(address common.Address, caller bind.ContractCaller) (*PrivatixServiceContractCaller, error) {
	contract, err := bindPrivatixServiceContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PrivatixServiceContractCaller{contract: contract}, nil
}

// NewPrivatixServiceContractTransactor creates a new write-only instance of PrivatixServiceContract, bound to a specific deployed contract.
func NewPrivatixServiceContractTransactor(address common.Address, transactor bind.ContractTransactor) (*PrivatixServiceContractTransactor, error) {
	contract, err := bindPrivatixServiceContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PrivatixServiceContractTransactor{contract: contract}, nil
}

// NewPrivatixServiceContractFilterer creates a new log filterer instance of PrivatixServiceContract, bound to a specific deployed contract.
func NewPrivatixServiceContractFilterer(address common.Address, filterer bind.ContractFilterer) (*PrivatixServiceContractFilterer, error) {
	contract, err := bindPrivatixServiceContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PrivatixServiceContractFilterer{contract: contract}, nil
}

// bindPrivatixServiceContract binds a generic wrapper to an already deployed contract.
func bindPrivatixServiceContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(PrivatixServiceContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PrivatixServiceContract *PrivatixServiceContractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PrivatixServiceContract.Contract.PrivatixServiceContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PrivatixServiceContract *PrivatixServiceContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.PrivatixServiceContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PrivatixServiceContract *PrivatixServiceContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.PrivatixServiceContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PrivatixServiceContract *PrivatixServiceContractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PrivatixServiceContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PrivatixServiceContract *PrivatixServiceContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PrivatixServiceContract *PrivatixServiceContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.contract.Transact(opts, method, params...)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_address address) constant returns(uint256)
func (_PrivatixServiceContract *PrivatixServiceContractCaller) BalanceOf(opts *bind.CallOpts, _address common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PrivatixServiceContract.contract.Call(opts, out, "balanceOf", _address)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_address address) constant returns(uint256)
func (_PrivatixServiceContract *PrivatixServiceContractSession) BalanceOf(_address common.Address) (*big.Int, error) {
	return _PrivatixServiceContract.Contract.BalanceOf(&_PrivatixServiceContract.CallOpts, _address)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_address address) constant returns(uint256)
func (_PrivatixServiceContract *PrivatixServiceContractCallerSession) BalanceOf(_address common.Address) (*big.Int, error) {
	return _PrivatixServiceContract.Contract.BalanceOf(&_PrivatixServiceContract.CallOpts, _address)
}

// Challenge_period is a free data retrieval call binding the contract method 0x0a00840c.
//
// Solidity: function challenge_period() constant returns(uint32)
func (_PrivatixServiceContract *PrivatixServiceContractCaller) Challenge_period(opts *bind.CallOpts) (uint32, error) {
	var (
		ret0 = new(uint32)
	)
	out := ret0
	err := _PrivatixServiceContract.contract.Call(opts, out, "challenge_period")
	return *ret0, err
}

// Challenge_period is a free data retrieval call binding the contract method 0x0a00840c.
//
// Solidity: function challenge_period() constant returns(uint32)
func (_PrivatixServiceContract *PrivatixServiceContractSession) Challenge_period() (uint32, error) {
	return _PrivatixServiceContract.Contract.Challenge_period(&_PrivatixServiceContract.CallOpts)
}

// Challenge_period is a free data retrieval call binding the contract method 0x0a00840c.
//
// Solidity: function challenge_period() constant returns(uint32)
func (_PrivatixServiceContract *PrivatixServiceContractCallerSession) Challenge_period() (uint32, error) {
	return _PrivatixServiceContract.Contract.Challenge_period(&_PrivatixServiceContract.CallOpts)
}

// Channel_deposit_bugbounty_limit is a free data retrieval call binding the contract method 0x6108b5ff.
//
// Solidity: function channel_deposit_bugbounty_limit() constant returns(uint256)
func (_PrivatixServiceContract *PrivatixServiceContractCaller) Channel_deposit_bugbounty_limit(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PrivatixServiceContract.contract.Call(opts, out, "channel_deposit_bugbounty_limit")
	return *ret0, err
}

// Channel_deposit_bugbounty_limit is a free data retrieval call binding the contract method 0x6108b5ff.
//
// Solidity: function channel_deposit_bugbounty_limit() constant returns(uint256)
func (_PrivatixServiceContract *PrivatixServiceContractSession) Channel_deposit_bugbounty_limit() (*big.Int, error) {
	return _PrivatixServiceContract.Contract.Channel_deposit_bugbounty_limit(&_PrivatixServiceContract.CallOpts)
}

// Channel_deposit_bugbounty_limit is a free data retrieval call binding the contract method 0x6108b5ff.
//
// Solidity: function channel_deposit_bugbounty_limit() constant returns(uint256)
func (_PrivatixServiceContract *PrivatixServiceContractCallerSession) Channel_deposit_bugbounty_limit() (*big.Int, error) {
	return _PrivatixServiceContract.Contract.Channel_deposit_bugbounty_limit(&_PrivatixServiceContract.CallOpts)
}

// Channels is a free data retrieval call binding the contract method 0x7a7ebd7b.
//
// Solidity: function channels( bytes32) constant returns(deposit uint192, open_block_number uint32)
func (_PrivatixServiceContract *PrivatixServiceContractCaller) Channels(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Deposit           *big.Int
	Open_block_number uint32
}, error) {
	ret := new(struct {
		Deposit           *big.Int
		Open_block_number uint32
	})
	out := ret
	err := _PrivatixServiceContract.contract.Call(opts, out, "channels", arg0)
	return *ret, err
}

// Channels is a free data retrieval call binding the contract method 0x7a7ebd7b.
//
// Solidity: function channels( bytes32) constant returns(deposit uint192, open_block_number uint32)
func (_PrivatixServiceContract *PrivatixServiceContractSession) Channels(arg0 [32]byte) (struct {
	Deposit           *big.Int
	Open_block_number uint32
}, error) {
	return _PrivatixServiceContract.Contract.Channels(&_PrivatixServiceContract.CallOpts, arg0)
}

// Channels is a free data retrieval call binding the contract method 0x7a7ebd7b.
//
// Solidity: function channels( bytes32) constant returns(deposit uint192, open_block_number uint32)
func (_PrivatixServiceContract *PrivatixServiceContractCallerSession) Channels(arg0 [32]byte) (struct {
	Deposit           *big.Int
	Open_block_number uint32
}, error) {
	return _PrivatixServiceContract.Contract.Channels(&_PrivatixServiceContract.CallOpts, arg0)
}

// Closing_requests is a free data retrieval call binding the contract method 0x77c13323.
//
// Solidity: function closing_requests( bytes32) constant returns(closing_balance uint192, settle_block_number uint32)
func (_PrivatixServiceContract *PrivatixServiceContractCaller) Closing_requests(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Closing_balance     *big.Int
	Settle_block_number uint32
}, error) {
	ret := new(struct {
		Closing_balance     *big.Int
		Settle_block_number uint32
	})
	out := ret
	err := _PrivatixServiceContract.contract.Call(opts, out, "closing_requests", arg0)
	return *ret, err
}

// Closing_requests is a free data retrieval call binding the contract method 0x77c13323.
//
// Solidity: function closing_requests( bytes32) constant returns(closing_balance uint192, settle_block_number uint32)
func (_PrivatixServiceContract *PrivatixServiceContractSession) Closing_requests(arg0 [32]byte) (struct {
	Closing_balance     *big.Int
	Settle_block_number uint32
}, error) {
	return _PrivatixServiceContract.Contract.Closing_requests(&_PrivatixServiceContract.CallOpts, arg0)
}

// Closing_requests is a free data retrieval call binding the contract method 0x77c13323.
//
// Solidity: function closing_requests( bytes32) constant returns(closing_balance uint192, settle_block_number uint32)
func (_PrivatixServiceContract *PrivatixServiceContractCallerSession) Closing_requests(arg0 [32]byte) (struct {
	Closing_balance     *big.Int
	Settle_block_number uint32
}, error) {
	return _PrivatixServiceContract.Contract.Closing_requests(&_PrivatixServiceContract.CallOpts, arg0)
}

// ExtractSignature is a free data retrieval call binding the contract method 0x4c4b3c41.
//
// Solidity: function extractSignature(_address address, _open_block_number uint32, _offering_hash bytes32, _balance uint192, _msg_sig bytes, _type bool) constant returns(address)
func (_PrivatixServiceContract *PrivatixServiceContractCaller) ExtractSignature(opts *bind.CallOpts, _address common.Address, _open_block_number uint32, _offering_hash [32]byte, _balance *big.Int, _msg_sig []byte, _type bool) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _PrivatixServiceContract.contract.Call(opts, out, "extractSignature", _address, _open_block_number, _offering_hash, _balance, _msg_sig, _type)
	return *ret0, err
}

// ExtractSignature is a free data retrieval call binding the contract method 0x4c4b3c41.
//
// Solidity: function extractSignature(_address address, _open_block_number uint32, _offering_hash bytes32, _balance uint192, _msg_sig bytes, _type bool) constant returns(address)
func (_PrivatixServiceContract *PrivatixServiceContractSession) ExtractSignature(_address common.Address, _open_block_number uint32, _offering_hash [32]byte, _balance *big.Int, _msg_sig []byte, _type bool) (common.Address, error) {
	return _PrivatixServiceContract.Contract.ExtractSignature(&_PrivatixServiceContract.CallOpts, _address, _open_block_number, _offering_hash, _balance, _msg_sig, _type)
}

// ExtractSignature is a free data retrieval call binding the contract method 0x4c4b3c41.
//
// Solidity: function extractSignature(_address address, _open_block_number uint32, _offering_hash bytes32, _balance uint192, _msg_sig bytes, _type bool) constant returns(address)
func (_PrivatixServiceContract *PrivatixServiceContractCallerSession) ExtractSignature(_address common.Address, _open_block_number uint32, _offering_hash [32]byte, _balance *big.Int, _msg_sig []byte, _type bool) (common.Address, error) {
	return _PrivatixServiceContract.Contract.ExtractSignature(&_PrivatixServiceContract.CallOpts, _address, _open_block_number, _offering_hash, _balance, _msg_sig, _type)
}

// GetChannelInfo is a free data retrieval call binding the contract method 0x52994187.
//
// Solidity: function getChannelInfo(_client_address address, _agent_address address, _open_block_number uint32, _offering_hash bytes32) constant returns(bytes32, uint192, uint32, uint192)
func (_PrivatixServiceContract *PrivatixServiceContractCaller) GetChannelInfo(opts *bind.CallOpts, _client_address common.Address, _agent_address common.Address, _open_block_number uint32, _offering_hash [32]byte) ([32]byte, *big.Int, uint32, *big.Int, error) {
	var (
		ret0 = new([32]byte)
		ret1 = new(*big.Int)
		ret2 = new(uint32)
		ret3 = new(*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
		ret3,
	}
	err := _PrivatixServiceContract.contract.Call(opts, out, "getChannelInfo", _client_address, _agent_address, _open_block_number, _offering_hash)
	return *ret0, *ret1, *ret2, *ret3, err
}

// GetChannelInfo is a free data retrieval call binding the contract method 0x52994187.
//
// Solidity: function getChannelInfo(_client_address address, _agent_address address, _open_block_number uint32, _offering_hash bytes32) constant returns(bytes32, uint192, uint32, uint192)
func (_PrivatixServiceContract *PrivatixServiceContractSession) GetChannelInfo(_client_address common.Address, _agent_address common.Address, _open_block_number uint32, _offering_hash [32]byte) ([32]byte, *big.Int, uint32, *big.Int, error) {
	return _PrivatixServiceContract.Contract.GetChannelInfo(&_PrivatixServiceContract.CallOpts, _client_address, _agent_address, _open_block_number, _offering_hash)
}

// GetChannelInfo is a free data retrieval call binding the contract method 0x52994187.
//
// Solidity: function getChannelInfo(_client_address address, _agent_address address, _open_block_number uint32, _offering_hash bytes32) constant returns(bytes32, uint192, uint32, uint192)
func (_PrivatixServiceContract *PrivatixServiceContractCallerSession) GetChannelInfo(_client_address common.Address, _agent_address common.Address, _open_block_number uint32, _offering_hash [32]byte) ([32]byte, *big.Int, uint32, *big.Int, error) {
	return _PrivatixServiceContract.Contract.GetChannelInfo(&_PrivatixServiceContract.CallOpts, _client_address, _agent_address, _open_block_number, _offering_hash)
}

// GetKey is a free data retrieval call binding the contract method 0x3fb56343.
//
// Solidity: function getKey(_client_address address, _agent_address address, _open_block_number uint32, _offering_hash bytes32) constant returns(data bytes32)
func (_PrivatixServiceContract *PrivatixServiceContractCaller) GetKey(opts *bind.CallOpts, _client_address common.Address, _agent_address common.Address, _open_block_number uint32, _offering_hash [32]byte) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _PrivatixServiceContract.contract.Call(opts, out, "getKey", _client_address, _agent_address, _open_block_number, _offering_hash)
	return *ret0, err
}

// GetKey is a free data retrieval call binding the contract method 0x3fb56343.
//
// Solidity: function getKey(_client_address address, _agent_address address, _open_block_number uint32, _offering_hash bytes32) constant returns(data bytes32)
func (_PrivatixServiceContract *PrivatixServiceContractSession) GetKey(_client_address common.Address, _agent_address common.Address, _open_block_number uint32, _offering_hash [32]byte) ([32]byte, error) {
	return _PrivatixServiceContract.Contract.GetKey(&_PrivatixServiceContract.CallOpts, _client_address, _agent_address, _open_block_number, _offering_hash)
}

// GetKey is a free data retrieval call binding the contract method 0x3fb56343.
//
// Solidity: function getKey(_client_address address, _agent_address address, _open_block_number uint32, _offering_hash bytes32) constant returns(data bytes32)
func (_PrivatixServiceContract *PrivatixServiceContractCallerSession) GetKey(_client_address common.Address, _agent_address common.Address, _open_block_number uint32, _offering_hash [32]byte) ([32]byte, error) {
	return _PrivatixServiceContract.Contract.GetKey(&_PrivatixServiceContract.CallOpts, _client_address, _agent_address, _open_block_number, _offering_hash)
}

// Internal_balances is a free data retrieval call binding the contract method 0x3884a42d.
//
// Solidity: function internal_balances( address) constant returns(uint192)
func (_PrivatixServiceContract *PrivatixServiceContractCaller) Internal_balances(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PrivatixServiceContract.contract.Call(opts, out, "internal_balances", arg0)
	return *ret0, err
}

// Internal_balances is a free data retrieval call binding the contract method 0x3884a42d.
//
// Solidity: function internal_balances( address) constant returns(uint192)
func (_PrivatixServiceContract *PrivatixServiceContractSession) Internal_balances(arg0 common.Address) (*big.Int, error) {
	return _PrivatixServiceContract.Contract.Internal_balances(&_PrivatixServiceContract.CallOpts, arg0)
}

// Internal_balances is a free data retrieval call binding the contract method 0x3884a42d.
//
// Solidity: function internal_balances( address) constant returns(uint192)
func (_PrivatixServiceContract *PrivatixServiceContractCallerSession) Internal_balances(arg0 common.Address) (*big.Int, error) {
	return _PrivatixServiceContract.Contract.Internal_balances(&_PrivatixServiceContract.CallOpts, arg0)
}

// Meta_version is a free data retrieval call binding the contract method 0xf1c23a76.
//
// Solidity: function meta_version() constant returns(string)
func (_PrivatixServiceContract *PrivatixServiceContractCaller) Meta_version(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _PrivatixServiceContract.contract.Call(opts, out, "meta_version")
	return *ret0, err
}

// Meta_version is a free data retrieval call binding the contract method 0xf1c23a76.
//
// Solidity: function meta_version() constant returns(string)
func (_PrivatixServiceContract *PrivatixServiceContractSession) Meta_version() (string, error) {
	return _PrivatixServiceContract.Contract.Meta_version(&_PrivatixServiceContract.CallOpts)
}

// Meta_version is a free data retrieval call binding the contract method 0xf1c23a76.
//
// Solidity: function meta_version() constant returns(string)
func (_PrivatixServiceContract *PrivatixServiceContractCallerSession) Meta_version() (string, error) {
	return _PrivatixServiceContract.Contract.Meta_version(&_PrivatixServiceContract.CallOpts)
}

// Network_fee is a free data retrieval call binding the contract method 0x66d4b643.
//
// Solidity: function network_fee() constant returns(uint32)
func (_PrivatixServiceContract *PrivatixServiceContractCaller) Network_fee(opts *bind.CallOpts) (uint32, error) {
	var (
		ret0 = new(uint32)
	)
	out := ret0
	err := _PrivatixServiceContract.contract.Call(opts, out, "network_fee")
	return *ret0, err
}

// Network_fee is a free data retrieval call binding the contract method 0x66d4b643.
//
// Solidity: function network_fee() constant returns(uint32)
func (_PrivatixServiceContract *PrivatixServiceContractSession) Network_fee() (uint32, error) {
	return _PrivatixServiceContract.Contract.Network_fee(&_PrivatixServiceContract.CallOpts)
}

// Network_fee is a free data retrieval call binding the contract method 0x66d4b643.
//
// Solidity: function network_fee() constant returns(uint32)
func (_PrivatixServiceContract *PrivatixServiceContractCallerSession) Network_fee() (uint32, error) {
	return _PrivatixServiceContract.Contract.Network_fee(&_PrivatixServiceContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_PrivatixServiceContract *PrivatixServiceContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _PrivatixServiceContract.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_PrivatixServiceContract *PrivatixServiceContractSession) Owner() (common.Address, error) {
	return _PrivatixServiceContract.Contract.Owner(&_PrivatixServiceContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_PrivatixServiceContract *PrivatixServiceContractCallerSession) Owner() (common.Address, error) {
	return _PrivatixServiceContract.Contract.Owner(&_PrivatixServiceContract.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() constant returns(address)
func (_PrivatixServiceContract *PrivatixServiceContractCaller) Token(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _PrivatixServiceContract.contract.Call(opts, out, "token")
	return *ret0, err
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() constant returns(address)
func (_PrivatixServiceContract *PrivatixServiceContractSession) Token() (common.Address, error) {
	return _PrivatixServiceContract.Contract.Token(&_PrivatixServiceContract.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() constant returns(address)
func (_PrivatixServiceContract *PrivatixServiceContractCallerSession) Token() (common.Address, error) {
	return _PrivatixServiceContract.Contract.Token(&_PrivatixServiceContract.CallOpts)
}

// AddBalanceERC20 is a paid mutator transaction binding the contract method 0x16b2ff10.
//
// Solidity: function addBalanceERC20(_value uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) AddBalanceERC20(opts *bind.TransactOpts, _value *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "addBalanceERC20", _value)
}

// AddBalanceERC20 is a paid mutator transaction binding the contract method 0x16b2ff10.
//
// Solidity: function addBalanceERC20(_value uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) AddBalanceERC20(_value *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.AddBalanceERC20(&_PrivatixServiceContract.TransactOpts, _value)
}

// AddBalanceERC20 is a paid mutator transaction binding the contract method 0x16b2ff10.
//
// Solidity: function addBalanceERC20(_value uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) AddBalanceERC20(_value *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.AddBalanceERC20(&_PrivatixServiceContract.TransactOpts, _value)
}

// CooperativeClose is a paid mutator transaction binding the contract method 0x0041e577.
//
// Solidity: function cooperativeClose(_agent_address address, _open_block_number uint32, _offering_hash bytes32, _balance uint192, _balance_msg_sig bytes, _closing_sig bytes) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) CooperativeClose(opts *bind.TransactOpts, _agent_address common.Address, _open_block_number uint32, _offering_hash [32]byte, _balance *big.Int, _balance_msg_sig []byte, _closing_sig []byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "cooperativeClose", _agent_address, _open_block_number, _offering_hash, _balance, _balance_msg_sig, _closing_sig)
}

// CooperativeClose is a paid mutator transaction binding the contract method 0x0041e577.
//
// Solidity: function cooperativeClose(_agent_address address, _open_block_number uint32, _offering_hash bytes32, _balance uint192, _balance_msg_sig bytes, _closing_sig bytes) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) CooperativeClose(_agent_address common.Address, _open_block_number uint32, _offering_hash [32]byte, _balance *big.Int, _balance_msg_sig []byte, _closing_sig []byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.CooperativeClose(&_PrivatixServiceContract.TransactOpts, _agent_address, _open_block_number, _offering_hash, _balance, _balance_msg_sig, _closing_sig)
}

// CooperativeClose is a paid mutator transaction binding the contract method 0x0041e577.
//
// Solidity: function cooperativeClose(_agent_address address, _open_block_number uint32, _offering_hash bytes32, _balance uint192, _balance_msg_sig bytes, _closing_sig bytes) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) CooperativeClose(_agent_address common.Address, _open_block_number uint32, _offering_hash [32]byte, _balance *big.Int, _balance_msg_sig []byte, _closing_sig []byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.CooperativeClose(&_PrivatixServiceContract.TransactOpts, _agent_address, _open_block_number, _offering_hash, _balance, _balance_msg_sig, _closing_sig)
}

// CreateChannel is a paid mutator transaction binding the contract method 0x6bc37152.
//
// Solidity: function createChannel(_agent_address address, _offering_hash bytes32, _deposit uint192, _authentication_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) CreateChannel(opts *bind.TransactOpts, _agent_address common.Address, _offering_hash [32]byte, _deposit *big.Int, _authentication_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "createChannel", _agent_address, _offering_hash, _deposit, _authentication_hash)
}

// CreateChannel is a paid mutator transaction binding the contract method 0x6bc37152.
//
// Solidity: function createChannel(_agent_address address, _offering_hash bytes32, _deposit uint192, _authentication_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) CreateChannel(_agent_address common.Address, _offering_hash [32]byte, _deposit *big.Int, _authentication_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.CreateChannel(&_PrivatixServiceContract.TransactOpts, _agent_address, _offering_hash, _deposit, _authentication_hash)
}

// CreateChannel is a paid mutator transaction binding the contract method 0x6bc37152.
//
// Solidity: function createChannel(_agent_address address, _offering_hash bytes32, _deposit uint192, _authentication_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) CreateChannel(_agent_address common.Address, _offering_hash [32]byte, _deposit *big.Int, _authentication_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.CreateChannel(&_PrivatixServiceContract.TransactOpts, _agent_address, _offering_hash, _deposit, _authentication_hash)
}

// PopupServiceOffering is a paid mutator transaction binding the contract method 0xb7ee7e6d.
//
// Solidity: function popupServiceOffering(_offering_hash bytes32) returns(success bool)
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) PopupServiceOffering(opts *bind.TransactOpts, _offering_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "popupServiceOffering", _offering_hash)
}

// PopupServiceOffering is a paid mutator transaction binding the contract method 0xb7ee7e6d.
//
// Solidity: function popupServiceOffering(_offering_hash bytes32) returns(success bool)
func (_PrivatixServiceContract *PrivatixServiceContractSession) PopupServiceOffering(_offering_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.PopupServiceOffering(&_PrivatixServiceContract.TransactOpts, _offering_hash)
}

// PopupServiceOffering is a paid mutator transaction binding the contract method 0xb7ee7e6d.
//
// Solidity: function popupServiceOffering(_offering_hash bytes32) returns(success bool)
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) PopupServiceOffering(_offering_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.PopupServiceOffering(&_PrivatixServiceContract.TransactOpts, _offering_hash)
}

// PublishServiceOfferingEndpoint is a paid mutator transaction binding the contract method 0x1ea64f00.
//
// Solidity: function publishServiceOfferingEndpoint(_client_address address, _offering_hash bytes32, _open_block_number uint32, _endpoint_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) PublishServiceOfferingEndpoint(opts *bind.TransactOpts, _client_address common.Address, _offering_hash [32]byte, _open_block_number uint32, _endpoint_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "publishServiceOfferingEndpoint", _client_address, _offering_hash, _open_block_number, _endpoint_hash)
}

// PublishServiceOfferingEndpoint is a paid mutator transaction binding the contract method 0x1ea64f00.
//
// Solidity: function publishServiceOfferingEndpoint(_client_address address, _offering_hash bytes32, _open_block_number uint32, _endpoint_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) PublishServiceOfferingEndpoint(_client_address common.Address, _offering_hash [32]byte, _open_block_number uint32, _endpoint_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.PublishServiceOfferingEndpoint(&_PrivatixServiceContract.TransactOpts, _client_address, _offering_hash, _open_block_number, _endpoint_hash)
}

// PublishServiceOfferingEndpoint is a paid mutator transaction binding the contract method 0x1ea64f00.
//
// Solidity: function publishServiceOfferingEndpoint(_client_address address, _offering_hash bytes32, _open_block_number uint32, _endpoint_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) PublishServiceOfferingEndpoint(_client_address common.Address, _offering_hash [32]byte, _open_block_number uint32, _endpoint_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.PublishServiceOfferingEndpoint(&_PrivatixServiceContract.TransactOpts, _client_address, _offering_hash, _open_block_number, _endpoint_hash)
}

// RegisterServiceOffering is a paid mutator transaction binding the contract method 0x68c6b821.
//
// Solidity: function registerServiceOffering(_offering_hash bytes32, _min_deposit uint192, _max_supply uint16) returns(success bool)
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) RegisterServiceOffering(opts *bind.TransactOpts, _offering_hash [32]byte, _min_deposit *big.Int, _max_supply uint16) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "registerServiceOffering", _offering_hash, _min_deposit, _max_supply)
}

// RegisterServiceOffering is a paid mutator transaction binding the contract method 0x68c6b821.
//
// Solidity: function registerServiceOffering(_offering_hash bytes32, _min_deposit uint192, _max_supply uint16) returns(success bool)
func (_PrivatixServiceContract *PrivatixServiceContractSession) RegisterServiceOffering(_offering_hash [32]byte, _min_deposit *big.Int, _max_supply uint16) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.RegisterServiceOffering(&_PrivatixServiceContract.TransactOpts, _offering_hash, _min_deposit, _max_supply)
}

// RegisterServiceOffering is a paid mutator transaction binding the contract method 0x68c6b821.
//
// Solidity: function registerServiceOffering(_offering_hash bytes32, _min_deposit uint192, _max_supply uint16) returns(success bool)
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) RegisterServiceOffering(_offering_hash [32]byte, _min_deposit *big.Int, _max_supply uint16) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.RegisterServiceOffering(&_PrivatixServiceContract.TransactOpts, _offering_hash, _min_deposit, _max_supply)
}

// RemoveServiceOffering is a paid mutator transaction binding the contract method 0xc09f8104.
//
// Solidity: function removeServiceOffering(_offering_hash bytes32) returns(success bool)
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) RemoveServiceOffering(opts *bind.TransactOpts, _offering_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "removeServiceOffering", _offering_hash)
}

// RemoveServiceOffering is a paid mutator transaction binding the contract method 0xc09f8104.
//
// Solidity: function removeServiceOffering(_offering_hash bytes32) returns(success bool)
func (_PrivatixServiceContract *PrivatixServiceContractSession) RemoveServiceOffering(_offering_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.RemoveServiceOffering(&_PrivatixServiceContract.TransactOpts, _offering_hash)
}

// RemoveServiceOffering is a paid mutator transaction binding the contract method 0xc09f8104.
//
// Solidity: function removeServiceOffering(_offering_hash bytes32) returns(success bool)
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) RemoveServiceOffering(_offering_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.RemoveServiceOffering(&_PrivatixServiceContract.TransactOpts, _offering_hash)
}

// ReturnBalanceERC20 is a paid mutator transaction binding the contract method 0x4f363052.
//
// Solidity: function returnBalanceERC20(_value uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) ReturnBalanceERC20(opts *bind.TransactOpts, _value *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "returnBalanceERC20", _value)
}

// ReturnBalanceERC20 is a paid mutator transaction binding the contract method 0x4f363052.
//
// Solidity: function returnBalanceERC20(_value uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) ReturnBalanceERC20(_value *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ReturnBalanceERC20(&_PrivatixServiceContract.TransactOpts, _value)
}

// ReturnBalanceERC20 is a paid mutator transaction binding the contract method 0x4f363052.
//
// Solidity: function returnBalanceERC20(_value uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) ReturnBalanceERC20(_value *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ReturnBalanceERC20(&_PrivatixServiceContract.TransactOpts, _value)
}

// SetNetworkFee is a paid mutator transaction binding the contract method 0x155bb726.
//
// Solidity: function setNetworkFee(_network_fee uint32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) SetNetworkFee(opts *bind.TransactOpts, _network_fee uint32) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "setNetworkFee", _network_fee)
}

// SetNetworkFee is a paid mutator transaction binding the contract method 0x155bb726.
//
// Solidity: function setNetworkFee(_network_fee uint32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) SetNetworkFee(_network_fee uint32) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.SetNetworkFee(&_PrivatixServiceContract.TransactOpts, _network_fee)
}

// SetNetworkFee is a paid mutator transaction binding the contract method 0x155bb726.
//
// Solidity: function setNetworkFee(_network_fee uint32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) SetNetworkFee(_network_fee uint32) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.SetNetworkFee(&_PrivatixServiceContract.TransactOpts, _network_fee)
}

// SetNetworkFeeAddress is a paid mutator transaction binding the contract method 0x5319022e.
//
// Solidity: function setNetworkFeeAddress(_network_fee_address address) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) SetNetworkFeeAddress(opts *bind.TransactOpts, _network_fee_address common.Address) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "setNetworkFeeAddress", _network_fee_address)
}

// SetNetworkFeeAddress is a paid mutator transaction binding the contract method 0x5319022e.
//
// Solidity: function setNetworkFeeAddress(_network_fee_address address) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) SetNetworkFeeAddress(_network_fee_address common.Address) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.SetNetworkFeeAddress(&_PrivatixServiceContract.TransactOpts, _network_fee_address)
}

// SetNetworkFeeAddress is a paid mutator transaction binding the contract method 0x5319022e.
//
// Solidity: function setNetworkFeeAddress(_network_fee_address address) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) SetNetworkFeeAddress(_network_fee_address common.Address) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.SetNetworkFeeAddress(&_PrivatixServiceContract.TransactOpts, _network_fee_address)
}

// Settle is a paid mutator transaction binding the contract method 0x27717c0b.
//
// Solidity: function settle(_agent_address address, _open_block_number uint32, _offering_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) Settle(opts *bind.TransactOpts, _agent_address common.Address, _open_block_number uint32, _offering_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "settle", _agent_address, _open_block_number, _offering_hash)
}

// Settle is a paid mutator transaction binding the contract method 0x27717c0b.
//
// Solidity: function settle(_agent_address address, _open_block_number uint32, _offering_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) Settle(_agent_address common.Address, _open_block_number uint32, _offering_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.Settle(&_PrivatixServiceContract.TransactOpts, _agent_address, _open_block_number, _offering_hash)
}

// Settle is a paid mutator transaction binding the contract method 0x27717c0b.
//
// Solidity: function settle(_agent_address address, _open_block_number uint32, _offering_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) Settle(_agent_address common.Address, _open_block_number uint32, _offering_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.Settle(&_PrivatixServiceContract.TransactOpts, _agent_address, _open_block_number, _offering_hash)
}

// ThrowEventLogChannelCloseRequested is a paid mutator transaction binding the contract method 0x92e24537.
//
// Solidity: function throwEventLogChannelCloseRequested(_agent_address address, _client_address address, _offering_hash bytes32, _open_block_number uint32, _balance uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) ThrowEventLogChannelCloseRequested(opts *bind.TransactOpts, _agent_address common.Address, _client_address common.Address, _offering_hash [32]byte, _open_block_number uint32, _balance *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "throwEventLogChannelCloseRequested", _agent_address, _client_address, _offering_hash, _open_block_number, _balance)
}

// ThrowEventLogChannelCloseRequested is a paid mutator transaction binding the contract method 0x92e24537.
//
// Solidity: function throwEventLogChannelCloseRequested(_agent_address address, _client_address address, _offering_hash bytes32, _open_block_number uint32, _balance uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) ThrowEventLogChannelCloseRequested(_agent_address common.Address, _client_address common.Address, _offering_hash [32]byte, _open_block_number uint32, _balance *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogChannelCloseRequested(&_PrivatixServiceContract.TransactOpts, _agent_address, _client_address, _offering_hash, _open_block_number, _balance)
}

// ThrowEventLogChannelCloseRequested is a paid mutator transaction binding the contract method 0x92e24537.
//
// Solidity: function throwEventLogChannelCloseRequested(_agent_address address, _client_address address, _offering_hash bytes32, _open_block_number uint32, _balance uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) ThrowEventLogChannelCloseRequested(_agent_address common.Address, _client_address common.Address, _offering_hash [32]byte, _open_block_number uint32, _balance *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogChannelCloseRequested(&_PrivatixServiceContract.TransactOpts, _agent_address, _client_address, _offering_hash, _open_block_number, _balance)
}

// ThrowEventLogChannelCreated is a paid mutator transaction binding the contract method 0x0a5ce4df.
//
// Solidity: function throwEventLogChannelCreated(_agent_address address, _client_address address, _offering_hash bytes32, _deposit uint192, _authentication_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) ThrowEventLogChannelCreated(opts *bind.TransactOpts, _agent_address common.Address, _client_address common.Address, _offering_hash [32]byte, _deposit *big.Int, _authentication_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "throwEventLogChannelCreated", _agent_address, _client_address, _offering_hash, _deposit, _authentication_hash)
}

// ThrowEventLogChannelCreated is a paid mutator transaction binding the contract method 0x0a5ce4df.
//
// Solidity: function throwEventLogChannelCreated(_agent_address address, _client_address address, _offering_hash bytes32, _deposit uint192, _authentication_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) ThrowEventLogChannelCreated(_agent_address common.Address, _client_address common.Address, _offering_hash [32]byte, _deposit *big.Int, _authentication_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogChannelCreated(&_PrivatixServiceContract.TransactOpts, _agent_address, _client_address, _offering_hash, _deposit, _authentication_hash)
}

// ThrowEventLogChannelCreated is a paid mutator transaction binding the contract method 0x0a5ce4df.
//
// Solidity: function throwEventLogChannelCreated(_agent_address address, _client_address address, _offering_hash bytes32, _deposit uint192, _authentication_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) ThrowEventLogChannelCreated(_agent_address common.Address, _client_address common.Address, _offering_hash [32]byte, _deposit *big.Int, _authentication_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogChannelCreated(&_PrivatixServiceContract.TransactOpts, _agent_address, _client_address, _offering_hash, _deposit, _authentication_hash)
}

// ThrowEventLogChannelToppedUp is a paid mutator transaction binding the contract method 0x4a6cb9d9.
//
// Solidity: function throwEventLogChannelToppedUp(_agent_address address, _client_address address, _offering_hash bytes32, _open_block_number uint32, _added_deposit uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) ThrowEventLogChannelToppedUp(opts *bind.TransactOpts, _agent_address common.Address, _client_address common.Address, _offering_hash [32]byte, _open_block_number uint32, _added_deposit *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "throwEventLogChannelToppedUp", _agent_address, _client_address, _offering_hash, _open_block_number, _added_deposit)
}

// ThrowEventLogChannelToppedUp is a paid mutator transaction binding the contract method 0x4a6cb9d9.
//
// Solidity: function throwEventLogChannelToppedUp(_agent_address address, _client_address address, _offering_hash bytes32, _open_block_number uint32, _added_deposit uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) ThrowEventLogChannelToppedUp(_agent_address common.Address, _client_address common.Address, _offering_hash [32]byte, _open_block_number uint32, _added_deposit *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogChannelToppedUp(&_PrivatixServiceContract.TransactOpts, _agent_address, _client_address, _offering_hash, _open_block_number, _added_deposit)
}

// ThrowEventLogChannelToppedUp is a paid mutator transaction binding the contract method 0x4a6cb9d9.
//
// Solidity: function throwEventLogChannelToppedUp(_agent_address address, _client_address address, _offering_hash bytes32, _open_block_number uint32, _added_deposit uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) ThrowEventLogChannelToppedUp(_agent_address common.Address, _client_address common.Address, _offering_hash [32]byte, _open_block_number uint32, _added_deposit *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogChannelToppedUp(&_PrivatixServiceContract.TransactOpts, _agent_address, _client_address, _offering_hash, _open_block_number, _added_deposit)
}

// ThrowEventLogCooperativeChannelClose is a paid mutator transaction binding the contract method 0xc0aa2dd3.
//
// Solidity: function throwEventLogCooperativeChannelClose(_agent_address address, _client_address address, _offering_hash bytes32, _open_block_number uint32, _balance uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) ThrowEventLogCooperativeChannelClose(opts *bind.TransactOpts, _agent_address common.Address, _client_address common.Address, _offering_hash [32]byte, _open_block_number uint32, _balance *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "throwEventLogCooperativeChannelClose", _agent_address, _client_address, _offering_hash, _open_block_number, _balance)
}

// ThrowEventLogCooperativeChannelClose is a paid mutator transaction binding the contract method 0xc0aa2dd3.
//
// Solidity: function throwEventLogCooperativeChannelClose(_agent_address address, _client_address address, _offering_hash bytes32, _open_block_number uint32, _balance uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) ThrowEventLogCooperativeChannelClose(_agent_address common.Address, _client_address common.Address, _offering_hash [32]byte, _open_block_number uint32, _balance *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogCooperativeChannelClose(&_PrivatixServiceContract.TransactOpts, _agent_address, _client_address, _offering_hash, _open_block_number, _balance)
}

// ThrowEventLogCooperativeChannelClose is a paid mutator transaction binding the contract method 0xc0aa2dd3.
//
// Solidity: function throwEventLogCooperativeChannelClose(_agent_address address, _client_address address, _offering_hash bytes32, _open_block_number uint32, _balance uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) ThrowEventLogCooperativeChannelClose(_agent_address common.Address, _client_address common.Address, _offering_hash [32]byte, _open_block_number uint32, _balance *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogCooperativeChannelClose(&_PrivatixServiceContract.TransactOpts, _agent_address, _client_address, _offering_hash, _open_block_number, _balance)
}

// ThrowEventLogOfferingCreated is a paid mutator transaction binding the contract method 0x6bbe1f35.
//
// Solidity: function throwEventLogOfferingCreated(_agent_address address, _offering_hash bytes32, _min_deposit uint256, _current_supply uint16) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) ThrowEventLogOfferingCreated(opts *bind.TransactOpts, _agent_address common.Address, _offering_hash [32]byte, _min_deposit *big.Int, _current_supply uint16) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "throwEventLogOfferingCreated", _agent_address, _offering_hash, _min_deposit, _current_supply)
}

// ThrowEventLogOfferingCreated is a paid mutator transaction binding the contract method 0x6bbe1f35.
//
// Solidity: function throwEventLogOfferingCreated(_agent_address address, _offering_hash bytes32, _min_deposit uint256, _current_supply uint16) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) ThrowEventLogOfferingCreated(_agent_address common.Address, _offering_hash [32]byte, _min_deposit *big.Int, _current_supply uint16) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogOfferingCreated(&_PrivatixServiceContract.TransactOpts, _agent_address, _offering_hash, _min_deposit, _current_supply)
}

// ThrowEventLogOfferingCreated is a paid mutator transaction binding the contract method 0x6bbe1f35.
//
// Solidity: function throwEventLogOfferingCreated(_agent_address address, _offering_hash bytes32, _min_deposit uint256, _current_supply uint16) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) ThrowEventLogOfferingCreated(_agent_address common.Address, _offering_hash [32]byte, _min_deposit *big.Int, _current_supply uint16) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogOfferingCreated(&_PrivatixServiceContract.TransactOpts, _agent_address, _offering_hash, _min_deposit, _current_supply)
}

// ThrowEventLogOfferingDeleted is a paid mutator transaction binding the contract method 0x2262a4ae.
//
// Solidity: function throwEventLogOfferingDeleted(_agent_address address, _offering_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) ThrowEventLogOfferingDeleted(opts *bind.TransactOpts, _agent_address common.Address, _offering_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "throwEventLogOfferingDeleted", _agent_address, _offering_hash)
}

// ThrowEventLogOfferingDeleted is a paid mutator transaction binding the contract method 0x2262a4ae.
//
// Solidity: function throwEventLogOfferingDeleted(_agent_address address, _offering_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) ThrowEventLogOfferingDeleted(_agent_address common.Address, _offering_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogOfferingDeleted(&_PrivatixServiceContract.TransactOpts, _agent_address, _offering_hash)
}

// ThrowEventLogOfferingDeleted is a paid mutator transaction binding the contract method 0x2262a4ae.
//
// Solidity: function throwEventLogOfferingDeleted(_agent_address address, _offering_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) ThrowEventLogOfferingDeleted(_agent_address common.Address, _offering_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogOfferingDeleted(&_PrivatixServiceContract.TransactOpts, _agent_address, _offering_hash)
}

// ThrowEventLogOfferingEndpoint is a paid mutator transaction binding the contract method 0x71a87244.
//
// Solidity: function throwEventLogOfferingEndpoint(_agent_address address, _client_address address, _offering_hash bytes32, _open_block_number uint32, _endpoint_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) ThrowEventLogOfferingEndpoint(opts *bind.TransactOpts, _agent_address common.Address, _client_address common.Address, _offering_hash [32]byte, _open_block_number uint32, _endpoint_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "throwEventLogOfferingEndpoint", _agent_address, _client_address, _offering_hash, _open_block_number, _endpoint_hash)
}

// ThrowEventLogOfferingEndpoint is a paid mutator transaction binding the contract method 0x71a87244.
//
// Solidity: function throwEventLogOfferingEndpoint(_agent_address address, _client_address address, _offering_hash bytes32, _open_block_number uint32, _endpoint_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) ThrowEventLogOfferingEndpoint(_agent_address common.Address, _client_address common.Address, _offering_hash [32]byte, _open_block_number uint32, _endpoint_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogOfferingEndpoint(&_PrivatixServiceContract.TransactOpts, _agent_address, _client_address, _offering_hash, _open_block_number, _endpoint_hash)
}

// ThrowEventLogOfferingEndpoint is a paid mutator transaction binding the contract method 0x71a87244.
//
// Solidity: function throwEventLogOfferingEndpoint(_agent_address address, _client_address address, _offering_hash bytes32, _open_block_number uint32, _endpoint_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) ThrowEventLogOfferingEndpoint(_agent_address common.Address, _client_address common.Address, _offering_hash [32]byte, _open_block_number uint32, _endpoint_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogOfferingEndpoint(&_PrivatixServiceContract.TransactOpts, _agent_address, _client_address, _offering_hash, _open_block_number, _endpoint_hash)
}

// ThrowEventLogOfferingPopedUp is a paid mutator transaction binding the contract method 0xfec5be78.
//
// Solidity: function throwEventLogOfferingPopedUp(_agent_address address, _offering_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) ThrowEventLogOfferingPopedUp(opts *bind.TransactOpts, _agent_address common.Address, _offering_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "throwEventLogOfferingPopedUp", _agent_address, _offering_hash)
}

// ThrowEventLogOfferingPopedUp is a paid mutator transaction binding the contract method 0xfec5be78.
//
// Solidity: function throwEventLogOfferingPopedUp(_agent_address address, _offering_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) ThrowEventLogOfferingPopedUp(_agent_address common.Address, _offering_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogOfferingPopedUp(&_PrivatixServiceContract.TransactOpts, _agent_address, _offering_hash)
}

// ThrowEventLogOfferingPopedUp is a paid mutator transaction binding the contract method 0xfec5be78.
//
// Solidity: function throwEventLogOfferingPopedUp(_agent_address address, _offering_hash bytes32) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) ThrowEventLogOfferingPopedUp(_agent_address common.Address, _offering_hash [32]byte) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogOfferingPopedUp(&_PrivatixServiceContract.TransactOpts, _agent_address, _offering_hash)
}

// ThrowEventLogOfferingSupplyChanged is a paid mutator transaction binding the contract method 0x8e1321cd.
//
// Solidity: function throwEventLogOfferingSupplyChanged(_agent_address address, _offering_hash bytes32, _current_supply uint16) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) ThrowEventLogOfferingSupplyChanged(opts *bind.TransactOpts, _agent_address common.Address, _offering_hash [32]byte, _current_supply uint16) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "throwEventLogOfferingSupplyChanged", _agent_address, _offering_hash, _current_supply)
}

// ThrowEventLogOfferingSupplyChanged is a paid mutator transaction binding the contract method 0x8e1321cd.
//
// Solidity: function throwEventLogOfferingSupplyChanged(_agent_address address, _offering_hash bytes32, _current_supply uint16) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) ThrowEventLogOfferingSupplyChanged(_agent_address common.Address, _offering_hash [32]byte, _current_supply uint16) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogOfferingSupplyChanged(&_PrivatixServiceContract.TransactOpts, _agent_address, _offering_hash, _current_supply)
}

// ThrowEventLogOfferingSupplyChanged is a paid mutator transaction binding the contract method 0x8e1321cd.
//
// Solidity: function throwEventLogOfferingSupplyChanged(_agent_address address, _offering_hash bytes32, _current_supply uint16) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) ThrowEventLogOfferingSupplyChanged(_agent_address common.Address, _offering_hash [32]byte, _current_supply uint16) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogOfferingSupplyChanged(&_PrivatixServiceContract.TransactOpts, _agent_address, _offering_hash, _current_supply)
}

// ThrowEventLogUnCooperativeChannelClose is a paid mutator transaction binding the contract method 0x3045e326.
//
// Solidity: function throwEventLogUnCooperativeChannelClose(_agent_address address, _client_address address, _offering_hash bytes32, _open_block_number uint32, _balance uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) ThrowEventLogUnCooperativeChannelClose(opts *bind.TransactOpts, _agent_address common.Address, _client_address common.Address, _offering_hash [32]byte, _open_block_number uint32, _balance *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "throwEventLogUnCooperativeChannelClose", _agent_address, _client_address, _offering_hash, _open_block_number, _balance)
}

// ThrowEventLogUnCooperativeChannelClose is a paid mutator transaction binding the contract method 0x3045e326.
//
// Solidity: function throwEventLogUnCooperativeChannelClose(_agent_address address, _client_address address, _offering_hash bytes32, _open_block_number uint32, _balance uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) ThrowEventLogUnCooperativeChannelClose(_agent_address common.Address, _client_address common.Address, _offering_hash [32]byte, _open_block_number uint32, _balance *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogUnCooperativeChannelClose(&_PrivatixServiceContract.TransactOpts, _agent_address, _client_address, _offering_hash, _open_block_number, _balance)
}

// ThrowEventLogUnCooperativeChannelClose is a paid mutator transaction binding the contract method 0x3045e326.
//
// Solidity: function throwEventLogUnCooperativeChannelClose(_agent_address address, _client_address address, _offering_hash bytes32, _open_block_number uint32, _balance uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) ThrowEventLogUnCooperativeChannelClose(_agent_address common.Address, _client_address common.Address, _offering_hash [32]byte, _open_block_number uint32, _balance *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.ThrowEventLogUnCooperativeChannelClose(&_PrivatixServiceContract.TransactOpts, _agent_address, _client_address, _offering_hash, _open_block_number, _balance)
}

// TopUpChannel is a paid mutator transaction binding the contract method 0xd4e2c042.
//
// Solidity: function topUpChannel(_agent_address address, _open_block_number uint32, _offering_hash bytes32, _added_deposit uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) TopUpChannel(opts *bind.TransactOpts, _agent_address common.Address, _open_block_number uint32, _offering_hash [32]byte, _added_deposit *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "topUpChannel", _agent_address, _open_block_number, _offering_hash, _added_deposit)
}

// TopUpChannel is a paid mutator transaction binding the contract method 0xd4e2c042.
//
// Solidity: function topUpChannel(_agent_address address, _open_block_number uint32, _offering_hash bytes32, _added_deposit uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) TopUpChannel(_agent_address common.Address, _open_block_number uint32, _offering_hash [32]byte, _added_deposit *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.TopUpChannel(&_PrivatixServiceContract.TransactOpts, _agent_address, _open_block_number, _offering_hash, _added_deposit)
}

// TopUpChannel is a paid mutator transaction binding the contract method 0xd4e2c042.
//
// Solidity: function topUpChannel(_agent_address address, _open_block_number uint32, _offering_hash bytes32, _added_deposit uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) TopUpChannel(_agent_address common.Address, _open_block_number uint32, _offering_hash [32]byte, _added_deposit *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.TopUpChannel(&_PrivatixServiceContract.TransactOpts, _agent_address, _open_block_number, _offering_hash, _added_deposit)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.TransferOwnership(&_PrivatixServiceContract.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(newOwner address) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.TransferOwnership(&_PrivatixServiceContract.TransactOpts, newOwner)
}

// UncooperativeClose is a paid mutator transaction binding the contract method 0xce05b6b4.
//
// Solidity: function uncooperativeClose(_agent_address address, _open_block_number uint32, _offering_hash bytes32, _balance uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactor) UncooperativeClose(opts *bind.TransactOpts, _agent_address common.Address, _open_block_number uint32, _offering_hash [32]byte, _balance *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.contract.Transact(opts, "uncooperativeClose", _agent_address, _open_block_number, _offering_hash, _balance)
}

// UncooperativeClose is a paid mutator transaction binding the contract method 0xce05b6b4.
//
// Solidity: function uncooperativeClose(_agent_address address, _open_block_number uint32, _offering_hash bytes32, _balance uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractSession) UncooperativeClose(_agent_address common.Address, _open_block_number uint32, _offering_hash [32]byte, _balance *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.UncooperativeClose(&_PrivatixServiceContract.TransactOpts, _agent_address, _open_block_number, _offering_hash, _balance)
}

// UncooperativeClose is a paid mutator transaction binding the contract method 0xce05b6b4.
//
// Solidity: function uncooperativeClose(_agent_address address, _open_block_number uint32, _offering_hash bytes32, _balance uint192) returns()
func (_PrivatixServiceContract *PrivatixServiceContractTransactorSession) UncooperativeClose(_agent_address common.Address, _open_block_number uint32, _offering_hash [32]byte, _balance *big.Int) (*types.Transaction, error) {
	return _PrivatixServiceContract.Contract.UncooperativeClose(&_PrivatixServiceContract.TransactOpts, _agent_address, _open_block_number, _offering_hash, _balance)
}

// PrivatixServiceContractLogChannelCloseRequestedIterator is returned from FilterLogChannelCloseRequested and is used to iterate over the raw logs and unpacked data for LogChannelCloseRequested events raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogChannelCloseRequestedIterator struct {
	Event *PrivatixServiceContractLogChannelCloseRequested // Event containing the contract specifics and raw log

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
func (it *PrivatixServiceContractLogChannelCloseRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PrivatixServiceContractLogChannelCloseRequested)
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
		it.Event = new(PrivatixServiceContractLogChannelCloseRequested)
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
func (it *PrivatixServiceContractLogChannelCloseRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PrivatixServiceContractLogChannelCloseRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PrivatixServiceContractLogChannelCloseRequested represents a LogChannelCloseRequested event raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogChannelCloseRequested struct {
	Agent             common.Address
	Client            common.Address
	Offering_hash     [32]byte
	Open_block_number uint32
	Balance           *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterLogChannelCloseRequested is a free log retrieval operation binding the contract event 0xb40564b1d36572b2942ad7cfc5a5a967f3ef08c82163a910dee760c5b629a32e.
//
// Solidity: event LogChannelCloseRequested(_agent indexed address, _client indexed address, _offering_hash indexed bytes32, _open_block_number uint32, _balance uint192)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) FilterLogChannelCloseRequested(opts *bind.FilterOpts, _agent []common.Address, _client []common.Address, _offering_hash [][32]byte) (*PrivatixServiceContractLogChannelCloseRequestedIterator, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _clientRule []interface{}
	for _, _clientItem := range _client {
		_clientRule = append(_clientRule, _clientItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.FilterLogs(opts, "LogChannelCloseRequested", _agentRule, _clientRule, _offering_hashRule)
	if err != nil {
		return nil, err
	}
	return &PrivatixServiceContractLogChannelCloseRequestedIterator{contract: _PrivatixServiceContract.contract, event: "LogChannelCloseRequested", logs: logs, sub: sub}, nil
}

// WatchLogChannelCloseRequested is a free log subscription operation binding the contract event 0xb40564b1d36572b2942ad7cfc5a5a967f3ef08c82163a910dee760c5b629a32e.
//
// Solidity: event LogChannelCloseRequested(_agent indexed address, _client indexed address, _offering_hash indexed bytes32, _open_block_number uint32, _balance uint192)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) WatchLogChannelCloseRequested(opts *bind.WatchOpts, sink chan<- *PrivatixServiceContractLogChannelCloseRequested, _agent []common.Address, _client []common.Address, _offering_hash [][32]byte) (event.Subscription, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _clientRule []interface{}
	for _, _clientItem := range _client {
		_clientRule = append(_clientRule, _clientItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.WatchLogs(opts, "LogChannelCloseRequested", _agentRule, _clientRule, _offering_hashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PrivatixServiceContractLogChannelCloseRequested)
				if err := _PrivatixServiceContract.contract.UnpackLog(event, "LogChannelCloseRequested", log); err != nil {
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

// PrivatixServiceContractLogChannelCreatedIterator is returned from FilterLogChannelCreated and is used to iterate over the raw logs and unpacked data for LogChannelCreated events raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogChannelCreatedIterator struct {
	Event *PrivatixServiceContractLogChannelCreated // Event containing the contract specifics and raw log

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
func (it *PrivatixServiceContractLogChannelCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PrivatixServiceContractLogChannelCreated)
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
		it.Event = new(PrivatixServiceContractLogChannelCreated)
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
func (it *PrivatixServiceContractLogChannelCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PrivatixServiceContractLogChannelCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PrivatixServiceContractLogChannelCreated represents a LogChannelCreated event raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogChannelCreated struct {
	Agent               common.Address
	Client              common.Address
	Offering_hash       [32]byte
	Deposit             *big.Int
	Authentication_hash [32]byte
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterLogChannelCreated is a free log retrieval operation binding the contract event 0xa6153987181667023837aee39c3f1a702a16e5e146323ef10fb96844a526143c.
//
// Solidity: event LogChannelCreated(_agent indexed address, _client indexed address, _offering_hash indexed bytes32, _deposit uint192, _authentication_hash bytes32)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) FilterLogChannelCreated(opts *bind.FilterOpts, _agent []common.Address, _client []common.Address, _offering_hash [][32]byte) (*PrivatixServiceContractLogChannelCreatedIterator, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _clientRule []interface{}
	for _, _clientItem := range _client {
		_clientRule = append(_clientRule, _clientItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.FilterLogs(opts, "LogChannelCreated", _agentRule, _clientRule, _offering_hashRule)
	if err != nil {
		return nil, err
	}
	return &PrivatixServiceContractLogChannelCreatedIterator{contract: _PrivatixServiceContract.contract, event: "LogChannelCreated", logs: logs, sub: sub}, nil
}

// WatchLogChannelCreated is a free log subscription operation binding the contract event 0xa6153987181667023837aee39c3f1a702a16e5e146323ef10fb96844a526143c.
//
// Solidity: event LogChannelCreated(_agent indexed address, _client indexed address, _offering_hash indexed bytes32, _deposit uint192, _authentication_hash bytes32)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) WatchLogChannelCreated(opts *bind.WatchOpts, sink chan<- *PrivatixServiceContractLogChannelCreated, _agent []common.Address, _client []common.Address, _offering_hash [][32]byte) (event.Subscription, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _clientRule []interface{}
	for _, _clientItem := range _client {
		_clientRule = append(_clientRule, _clientItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.WatchLogs(opts, "LogChannelCreated", _agentRule, _clientRule, _offering_hashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PrivatixServiceContractLogChannelCreated)
				if err := _PrivatixServiceContract.contract.UnpackLog(event, "LogChannelCreated", log); err != nil {
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

// PrivatixServiceContractLogChannelToppedUpIterator is returned from FilterLogChannelToppedUp and is used to iterate over the raw logs and unpacked data for LogChannelToppedUp events raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogChannelToppedUpIterator struct {
	Event *PrivatixServiceContractLogChannelToppedUp // Event containing the contract specifics and raw log

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
func (it *PrivatixServiceContractLogChannelToppedUpIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PrivatixServiceContractLogChannelToppedUp)
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
		it.Event = new(PrivatixServiceContractLogChannelToppedUp)
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
func (it *PrivatixServiceContractLogChannelToppedUpIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PrivatixServiceContractLogChannelToppedUpIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PrivatixServiceContractLogChannelToppedUp represents a LogChannelToppedUp event raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogChannelToppedUp struct {
	Agent             common.Address
	Client            common.Address
	Offering_hash     [32]byte
	Open_block_number uint32
	Added_deposit     *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterLogChannelToppedUp is a free log retrieval operation binding the contract event 0xa3b2cd532a9050531ecc674928d7704894707ede1a436bfbee86b96b83f2a5ce.
//
// Solidity: event LogChannelToppedUp(_agent indexed address, _client indexed address, _offering_hash indexed bytes32, _open_block_number uint32, _added_deposit uint192)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) FilterLogChannelToppedUp(opts *bind.FilterOpts, _agent []common.Address, _client []common.Address, _offering_hash [][32]byte) (*PrivatixServiceContractLogChannelToppedUpIterator, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _clientRule []interface{}
	for _, _clientItem := range _client {
		_clientRule = append(_clientRule, _clientItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.FilterLogs(opts, "LogChannelToppedUp", _agentRule, _clientRule, _offering_hashRule)
	if err != nil {
		return nil, err
	}
	return &PrivatixServiceContractLogChannelToppedUpIterator{contract: _PrivatixServiceContract.contract, event: "LogChannelToppedUp", logs: logs, sub: sub}, nil
}

// WatchLogChannelToppedUp is a free log subscription operation binding the contract event 0xa3b2cd532a9050531ecc674928d7704894707ede1a436bfbee86b96b83f2a5ce.
//
// Solidity: event LogChannelToppedUp(_agent indexed address, _client indexed address, _offering_hash indexed bytes32, _open_block_number uint32, _added_deposit uint192)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) WatchLogChannelToppedUp(opts *bind.WatchOpts, sink chan<- *PrivatixServiceContractLogChannelToppedUp, _agent []common.Address, _client []common.Address, _offering_hash [][32]byte) (event.Subscription, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _clientRule []interface{}
	for _, _clientItem := range _client {
		_clientRule = append(_clientRule, _clientItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.WatchLogs(opts, "LogChannelToppedUp", _agentRule, _clientRule, _offering_hashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PrivatixServiceContractLogChannelToppedUp)
				if err := _PrivatixServiceContract.contract.UnpackLog(event, "LogChannelToppedUp", log); err != nil {
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

// PrivatixServiceContractLogCooperativeChannelCloseIterator is returned from FilterLogCooperativeChannelClose and is used to iterate over the raw logs and unpacked data for LogCooperativeChannelClose events raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogCooperativeChannelCloseIterator struct {
	Event *PrivatixServiceContractLogCooperativeChannelClose // Event containing the contract specifics and raw log

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
func (it *PrivatixServiceContractLogCooperativeChannelCloseIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PrivatixServiceContractLogCooperativeChannelClose)
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
		it.Event = new(PrivatixServiceContractLogCooperativeChannelClose)
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
func (it *PrivatixServiceContractLogCooperativeChannelCloseIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PrivatixServiceContractLogCooperativeChannelCloseIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PrivatixServiceContractLogCooperativeChannelClose represents a LogCooperativeChannelClose event raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogCooperativeChannelClose struct {
	Agent             common.Address
	Client            common.Address
	Offering_hash     [32]byte
	Open_block_number uint32
	Balance           *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterLogCooperativeChannelClose is a free log retrieval operation binding the contract event 0xb488ea0f49970f556cf18e57588e78dcc1d3fd45c71130aa5099a79e8b06c8e7.
//
// Solidity: event LogCooperativeChannelClose(_agent indexed address, _client indexed address, _offering_hash indexed bytes32, _open_block_number uint32, _balance uint192)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) FilterLogCooperativeChannelClose(opts *bind.FilterOpts, _agent []common.Address, _client []common.Address, _offering_hash [][32]byte) (*PrivatixServiceContractLogCooperativeChannelCloseIterator, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _clientRule []interface{}
	for _, _clientItem := range _client {
		_clientRule = append(_clientRule, _clientItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.FilterLogs(opts, "LogCooperativeChannelClose", _agentRule, _clientRule, _offering_hashRule)
	if err != nil {
		return nil, err
	}
	return &PrivatixServiceContractLogCooperativeChannelCloseIterator{contract: _PrivatixServiceContract.contract, event: "LogCooperativeChannelClose", logs: logs, sub: sub}, nil
}

// WatchLogCooperativeChannelClose is a free log subscription operation binding the contract event 0xb488ea0f49970f556cf18e57588e78dcc1d3fd45c71130aa5099a79e8b06c8e7.
//
// Solidity: event LogCooperativeChannelClose(_agent indexed address, _client indexed address, _offering_hash indexed bytes32, _open_block_number uint32, _balance uint192)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) WatchLogCooperativeChannelClose(opts *bind.WatchOpts, sink chan<- *PrivatixServiceContractLogCooperativeChannelClose, _agent []common.Address, _client []common.Address, _offering_hash [][32]byte) (event.Subscription, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _clientRule []interface{}
	for _, _clientItem := range _client {
		_clientRule = append(_clientRule, _clientItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.WatchLogs(opts, "LogCooperativeChannelClose", _agentRule, _clientRule, _offering_hashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PrivatixServiceContractLogCooperativeChannelClose)
				if err := _PrivatixServiceContract.contract.UnpackLog(event, "LogCooperativeChannelClose", log); err != nil {
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

// PrivatixServiceContractLogOfferingCreatedIterator is returned from FilterLogOfferingCreated and is used to iterate over the raw logs and unpacked data for LogOfferingCreated events raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogOfferingCreatedIterator struct {
	Event *PrivatixServiceContractLogOfferingCreated // Event containing the contract specifics and raw log

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
func (it *PrivatixServiceContractLogOfferingCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PrivatixServiceContractLogOfferingCreated)
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
		it.Event = new(PrivatixServiceContractLogOfferingCreated)
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
func (it *PrivatixServiceContractLogOfferingCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PrivatixServiceContractLogOfferingCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PrivatixServiceContractLogOfferingCreated represents a LogOfferingCreated event raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogOfferingCreated struct {
	Agent          common.Address
	Offering_hash  [32]byte
	Min_deposit    *big.Int
	Current_supply uint16
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterLogOfferingCreated is a free log retrieval operation binding the contract event 0x32c1913dfde418197923027c2f2260f19903a2e86a93ed83c4689ac91a96bafd.
//
// Solidity: event LogOfferingCreated(_agent indexed address, _offering_hash indexed bytes32, _min_deposit indexed uint256, _current_supply uint16)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) FilterLogOfferingCreated(opts *bind.FilterOpts, _agent []common.Address, _offering_hash [][32]byte, _min_deposit []*big.Int) (*PrivatixServiceContractLogOfferingCreatedIterator, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}
	var _min_depositRule []interface{}
	for _, _min_depositItem := range _min_deposit {
		_min_depositRule = append(_min_depositRule, _min_depositItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.FilterLogs(opts, "LogOfferingCreated", _agentRule, _offering_hashRule, _min_depositRule)
	if err != nil {
		return nil, err
	}
	return &PrivatixServiceContractLogOfferingCreatedIterator{contract: _PrivatixServiceContract.contract, event: "LogOfferingCreated", logs: logs, sub: sub}, nil
}

// WatchLogOfferingCreated is a free log subscription operation binding the contract event 0x32c1913dfde418197923027c2f2260f19903a2e86a93ed83c4689ac91a96bafd.
//
// Solidity: event LogOfferingCreated(_agent indexed address, _offering_hash indexed bytes32, _min_deposit indexed uint256, _current_supply uint16)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) WatchLogOfferingCreated(opts *bind.WatchOpts, sink chan<- *PrivatixServiceContractLogOfferingCreated, _agent []common.Address, _offering_hash [][32]byte, _min_deposit []*big.Int) (event.Subscription, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}
	var _min_depositRule []interface{}
	for _, _min_depositItem := range _min_deposit {
		_min_depositRule = append(_min_depositRule, _min_depositItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.WatchLogs(opts, "LogOfferingCreated", _agentRule, _offering_hashRule, _min_depositRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PrivatixServiceContractLogOfferingCreated)
				if err := _PrivatixServiceContract.contract.UnpackLog(event, "LogOfferingCreated", log); err != nil {
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

// PrivatixServiceContractLogOfferingDeletedIterator is returned from FilterLogOfferingDeleted and is used to iterate over the raw logs and unpacked data for LogOfferingDeleted events raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogOfferingDeletedIterator struct {
	Event *PrivatixServiceContractLogOfferingDeleted // Event containing the contract specifics and raw log

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
func (it *PrivatixServiceContractLogOfferingDeletedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PrivatixServiceContractLogOfferingDeleted)
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
		it.Event = new(PrivatixServiceContractLogOfferingDeleted)
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
func (it *PrivatixServiceContractLogOfferingDeletedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PrivatixServiceContractLogOfferingDeletedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PrivatixServiceContractLogOfferingDeleted represents a LogOfferingDeleted event raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogOfferingDeleted struct {
	Agent         common.Address
	Offering_hash [32]byte
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterLogOfferingDeleted is a free log retrieval operation binding the contract event 0xc3013cd9dd5c33b95a9cc1bc076481c9a6a1970be6d7f1ed33adafad6e57d3d6.
//
// Solidity: event LogOfferingDeleted(_agent indexed address, _offering_hash indexed bytes32)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) FilterLogOfferingDeleted(opts *bind.FilterOpts, _agent []common.Address, _offering_hash [][32]byte) (*PrivatixServiceContractLogOfferingDeletedIterator, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.FilterLogs(opts, "LogOfferingDeleted", _agentRule, _offering_hashRule)
	if err != nil {
		return nil, err
	}
	return &PrivatixServiceContractLogOfferingDeletedIterator{contract: _PrivatixServiceContract.contract, event: "LogOfferingDeleted", logs: logs, sub: sub}, nil
}

// WatchLogOfferingDeleted is a free log subscription operation binding the contract event 0xc3013cd9dd5c33b95a9cc1bc076481c9a6a1970be6d7f1ed33adafad6e57d3d6.
//
// Solidity: event LogOfferingDeleted(_agent indexed address, _offering_hash indexed bytes32)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) WatchLogOfferingDeleted(opts *bind.WatchOpts, sink chan<- *PrivatixServiceContractLogOfferingDeleted, _agent []common.Address, _offering_hash [][32]byte) (event.Subscription, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.WatchLogs(opts, "LogOfferingDeleted", _agentRule, _offering_hashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PrivatixServiceContractLogOfferingDeleted)
				if err := _PrivatixServiceContract.contract.UnpackLog(event, "LogOfferingDeleted", log); err != nil {
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

// PrivatixServiceContractLogOfferingEndpointIterator is returned from FilterLogOfferingEndpoint and is used to iterate over the raw logs and unpacked data for LogOfferingEndpoint events raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogOfferingEndpointIterator struct {
	Event *PrivatixServiceContractLogOfferingEndpoint // Event containing the contract specifics and raw log

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
func (it *PrivatixServiceContractLogOfferingEndpointIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PrivatixServiceContractLogOfferingEndpoint)
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
		it.Event = new(PrivatixServiceContractLogOfferingEndpoint)
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
func (it *PrivatixServiceContractLogOfferingEndpointIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PrivatixServiceContractLogOfferingEndpointIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PrivatixServiceContractLogOfferingEndpoint represents a LogOfferingEndpoint event raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogOfferingEndpoint struct {
	Agent             common.Address
	Client            common.Address
	Offering_hash     [32]byte
	Open_block_number uint32
	Endpoint_hash     [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterLogOfferingEndpoint is a free log retrieval operation binding the contract event 0x450e7ab61f9e1c40dd7c79edcba274a7a96f025fab1733b3fa1087a1b5d1db7d.
//
// Solidity: event LogOfferingEndpoint(_agent indexed address, _client indexed address, _offering_hash indexed bytes32, _open_block_number uint32, _endpoint_hash bytes32)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) FilterLogOfferingEndpoint(opts *bind.FilterOpts, _agent []common.Address, _client []common.Address, _offering_hash [][32]byte) (*PrivatixServiceContractLogOfferingEndpointIterator, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _clientRule []interface{}
	for _, _clientItem := range _client {
		_clientRule = append(_clientRule, _clientItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.FilterLogs(opts, "LogOfferingEndpoint", _agentRule, _clientRule, _offering_hashRule)
	if err != nil {
		return nil, err
	}
	return &PrivatixServiceContractLogOfferingEndpointIterator{contract: _PrivatixServiceContract.contract, event: "LogOfferingEndpoint", logs: logs, sub: sub}, nil
}

// WatchLogOfferingEndpoint is a free log subscription operation binding the contract event 0x450e7ab61f9e1c40dd7c79edcba274a7a96f025fab1733b3fa1087a1b5d1db7d.
//
// Solidity: event LogOfferingEndpoint(_agent indexed address, _client indexed address, _offering_hash indexed bytes32, _open_block_number uint32, _endpoint_hash bytes32)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) WatchLogOfferingEndpoint(opts *bind.WatchOpts, sink chan<- *PrivatixServiceContractLogOfferingEndpoint, _agent []common.Address, _client []common.Address, _offering_hash [][32]byte) (event.Subscription, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _clientRule []interface{}
	for _, _clientItem := range _client {
		_clientRule = append(_clientRule, _clientItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.WatchLogs(opts, "LogOfferingEndpoint", _agentRule, _clientRule, _offering_hashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PrivatixServiceContractLogOfferingEndpoint)
				if err := _PrivatixServiceContract.contract.UnpackLog(event, "LogOfferingEndpoint", log); err != nil {
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

// PrivatixServiceContractLogOfferingPopedUpIterator is returned from FilterLogOfferingPopedUp and is used to iterate over the raw logs and unpacked data for LogOfferingPopedUp events raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogOfferingPopedUpIterator struct {
	Event *PrivatixServiceContractLogOfferingPopedUp // Event containing the contract specifics and raw log

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
func (it *PrivatixServiceContractLogOfferingPopedUpIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PrivatixServiceContractLogOfferingPopedUp)
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
		it.Event = new(PrivatixServiceContractLogOfferingPopedUp)
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
func (it *PrivatixServiceContractLogOfferingPopedUpIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PrivatixServiceContractLogOfferingPopedUpIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PrivatixServiceContractLogOfferingPopedUp represents a LogOfferingPopedUp event raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogOfferingPopedUp struct {
	Agent         common.Address
	Offering_hash [32]byte
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterLogOfferingPopedUp is a free log retrieval operation binding the contract event 0xc37352067a3ca1eafcf2dc5ba537fc473509c4e4aaca729cb1dab7053ec1ffbf.
//
// Solidity: event LogOfferingPopedUp(_agent indexed address, _offering_hash indexed bytes32)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) FilterLogOfferingPopedUp(opts *bind.FilterOpts, _agent []common.Address, _offering_hash [][32]byte) (*PrivatixServiceContractLogOfferingPopedUpIterator, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.FilterLogs(opts, "LogOfferingPopedUp", _agentRule, _offering_hashRule)
	if err != nil {
		return nil, err
	}
	return &PrivatixServiceContractLogOfferingPopedUpIterator{contract: _PrivatixServiceContract.contract, event: "LogOfferingPopedUp", logs: logs, sub: sub}, nil
}

// WatchLogOfferingPopedUp is a free log subscription operation binding the contract event 0xc37352067a3ca1eafcf2dc5ba537fc473509c4e4aaca729cb1dab7053ec1ffbf.
//
// Solidity: event LogOfferingPopedUp(_agent indexed address, _offering_hash indexed bytes32)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) WatchLogOfferingPopedUp(opts *bind.WatchOpts, sink chan<- *PrivatixServiceContractLogOfferingPopedUp, _agent []common.Address, _offering_hash [][32]byte) (event.Subscription, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.WatchLogs(opts, "LogOfferingPopedUp", _agentRule, _offering_hashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PrivatixServiceContractLogOfferingPopedUp)
				if err := _PrivatixServiceContract.contract.UnpackLog(event, "LogOfferingPopedUp", log); err != nil {
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

// PrivatixServiceContractLogOfferingSupplyChangedIterator is returned from FilterLogOfferingSupplyChanged and is used to iterate over the raw logs and unpacked data for LogOfferingSupplyChanged events raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogOfferingSupplyChangedIterator struct {
	Event *PrivatixServiceContractLogOfferingSupplyChanged // Event containing the contract specifics and raw log

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
func (it *PrivatixServiceContractLogOfferingSupplyChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PrivatixServiceContractLogOfferingSupplyChanged)
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
		it.Event = new(PrivatixServiceContractLogOfferingSupplyChanged)
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
func (it *PrivatixServiceContractLogOfferingSupplyChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PrivatixServiceContractLogOfferingSupplyChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PrivatixServiceContractLogOfferingSupplyChanged represents a LogOfferingSupplyChanged event raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogOfferingSupplyChanged struct {
	Agent          common.Address
	Offering_hash  [32]byte
	Current_supply uint16
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterLogOfferingSupplyChanged is a free log retrieval operation binding the contract event 0x5848091ba8411ca73c2b3bcfa5ffdcc4db482c6bf114dfa984f75c03dd878cf3.
//
// Solidity: event LogOfferingSupplyChanged(_agent indexed address, _offering_hash indexed bytes32, _current_supply uint16)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) FilterLogOfferingSupplyChanged(opts *bind.FilterOpts, _agent []common.Address, _offering_hash [][32]byte) (*PrivatixServiceContractLogOfferingSupplyChangedIterator, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.FilterLogs(opts, "LogOfferingSupplyChanged", _agentRule, _offering_hashRule)
	if err != nil {
		return nil, err
	}
	return &PrivatixServiceContractLogOfferingSupplyChangedIterator{contract: _PrivatixServiceContract.contract, event: "LogOfferingSupplyChanged", logs: logs, sub: sub}, nil
}

// WatchLogOfferingSupplyChanged is a free log subscription operation binding the contract event 0x5848091ba8411ca73c2b3bcfa5ffdcc4db482c6bf114dfa984f75c03dd878cf3.
//
// Solidity: event LogOfferingSupplyChanged(_agent indexed address, _offering_hash indexed bytes32, _current_supply uint16)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) WatchLogOfferingSupplyChanged(opts *bind.WatchOpts, sink chan<- *PrivatixServiceContractLogOfferingSupplyChanged, _agent []common.Address, _offering_hash [][32]byte) (event.Subscription, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.WatchLogs(opts, "LogOfferingSupplyChanged", _agentRule, _offering_hashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PrivatixServiceContractLogOfferingSupplyChanged)
				if err := _PrivatixServiceContract.contract.UnpackLog(event, "LogOfferingSupplyChanged", log); err != nil {
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

// PrivatixServiceContractLogUnCooperativeChannelCloseIterator is returned from FilterLogUnCooperativeChannelClose and is used to iterate over the raw logs and unpacked data for LogUnCooperativeChannelClose events raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogUnCooperativeChannelCloseIterator struct {
	Event *PrivatixServiceContractLogUnCooperativeChannelClose // Event containing the contract specifics and raw log

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
func (it *PrivatixServiceContractLogUnCooperativeChannelCloseIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PrivatixServiceContractLogUnCooperativeChannelClose)
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
		it.Event = new(PrivatixServiceContractLogUnCooperativeChannelClose)
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
func (it *PrivatixServiceContractLogUnCooperativeChannelCloseIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PrivatixServiceContractLogUnCooperativeChannelCloseIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PrivatixServiceContractLogUnCooperativeChannelClose represents a LogUnCooperativeChannelClose event raised by the PrivatixServiceContract contract.
type PrivatixServiceContractLogUnCooperativeChannelClose struct {
	Agent             common.Address
	Client            common.Address
	Offering_hash     [32]byte
	Open_block_number uint32
	Balance           *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterLogUnCooperativeChannelClose is a free log retrieval operation binding the contract event 0x7418f9b30b6de272d9d54ee6822f674042c58cea183b76d5d4e7b3c933a158f6.
//
// Solidity: event LogUnCooperativeChannelClose(_agent indexed address, _client indexed address, _offering_hash indexed bytes32, _open_block_number uint32, _balance uint192)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) FilterLogUnCooperativeChannelClose(opts *bind.FilterOpts, _agent []common.Address, _client []common.Address, _offering_hash [][32]byte) (*PrivatixServiceContractLogUnCooperativeChannelCloseIterator, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _clientRule []interface{}
	for _, _clientItem := range _client {
		_clientRule = append(_clientRule, _clientItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.FilterLogs(opts, "LogUnCooperativeChannelClose", _agentRule, _clientRule, _offering_hashRule)
	if err != nil {
		return nil, err
	}
	return &PrivatixServiceContractLogUnCooperativeChannelCloseIterator{contract: _PrivatixServiceContract.contract, event: "LogUnCooperativeChannelClose", logs: logs, sub: sub}, nil
}

// WatchLogUnCooperativeChannelClose is a free log subscription operation binding the contract event 0x7418f9b30b6de272d9d54ee6822f674042c58cea183b76d5d4e7b3c933a158f6.
//
// Solidity: event LogUnCooperativeChannelClose(_agent indexed address, _client indexed address, _offering_hash indexed bytes32, _open_block_number uint32, _balance uint192)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) WatchLogUnCooperativeChannelClose(opts *bind.WatchOpts, sink chan<- *PrivatixServiceContractLogUnCooperativeChannelClose, _agent []common.Address, _client []common.Address, _offering_hash [][32]byte) (event.Subscription, error) {

	var _agentRule []interface{}
	for _, _agentItem := range _agent {
		_agentRule = append(_agentRule, _agentItem)
	}
	var _clientRule []interface{}
	for _, _clientItem := range _client {
		_clientRule = append(_clientRule, _clientItem)
	}
	var _offering_hashRule []interface{}
	for _, _offering_hashItem := range _offering_hash {
		_offering_hashRule = append(_offering_hashRule, _offering_hashItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.WatchLogs(opts, "LogUnCooperativeChannelClose", _agentRule, _clientRule, _offering_hashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PrivatixServiceContractLogUnCooperativeChannelClose)
				if err := _PrivatixServiceContract.contract.UnpackLog(event, "LogUnCooperativeChannelClose", log); err != nil {
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

// PrivatixServiceContractOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the PrivatixServiceContract contract.
type PrivatixServiceContractOwnershipTransferredIterator struct {
	Event *PrivatixServiceContractOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *PrivatixServiceContractOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PrivatixServiceContractOwnershipTransferred)
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
		it.Event = new(PrivatixServiceContractOwnershipTransferred)
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
func (it *PrivatixServiceContractOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PrivatixServiceContractOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PrivatixServiceContractOwnershipTransferred represents a OwnershipTransferred event raised by the PrivatixServiceContract contract.
type PrivatixServiceContractOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*PrivatixServiceContractOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &PrivatixServiceContractOwnershipTransferredIterator{contract: _PrivatixServiceContract.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(previousOwner indexed address, newOwner indexed address)
func (_PrivatixServiceContract *PrivatixServiceContractFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *PrivatixServiceContractOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _PrivatixServiceContract.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PrivatixServiceContractOwnershipTransferred)
				if err := _PrivatixServiceContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
