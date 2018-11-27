package eth

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/privatix/dappctrl/data"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/util/log"
)

const (
	second = time.Second
)

// Backend adapter to communicate with contract.
type Backend interface {
	LatestBlockNumber(ctx context.Context) (*big.Int, error)

	SuggestGasPrice(ctx context.Context) (*big.Int, error)

	EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error)

	CooperativeClose(*bind.TransactOpts, common.Address, uint32,
		[common.HashLength]byte, *big.Int, []byte, []byte) (*types.Transaction, error)

	GetTransactionByHash(context.Context, common.Hash) (*types.Transaction, bool, error)

	RegisterServiceOffering(*bind.TransactOpts, [common.HashLength]byte,
		*big.Int, uint16, uint8, data.Base64String) (*types.Transaction, error)

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

	PSCPopupServiceOffering(opts *bind.TransactOpts, offeringHash [32]byte,
		somcType uint8, somcData data.Base64String) (*types.Transaction, error)

	FilterLogs(ctx context.Context,
		q ethereum.FilterQuery) ([]types.Log, error)

	HeaderByNumber(ctx context.Context,
		number *big.Int) (*types.Header, error)

	PTCAddress() common.Address

	PSCAddress() common.Address
}

type backendInstance struct {
	cfg    *Config
	psc    *contract.PrivatixServiceContract
	ptc    *contract.PrivatixTokenContract
	conn   *client
	logger log.Logger
}

// NewBackend returns eth back implementation.
func NewBackend(cfg *Config, logger log.Logger) Backend {
	conn, ptc, psc, err := newInstance(cfg, logger)
	if err != nil {
		logger.Fatal(err.Error())
	}

	b := &backendInstance{cfg: cfg, ptc: ptc, psc: psc,
		conn: conn, logger: logger,
	}

	go b.connectionControl()

	return b
}

func newInstance(cfg *Config,
	logger log.Logger) (*client, *contract.PrivatixTokenContract,
	*contract.PrivatixServiceContract, error) {
	conn, err := newClient(cfg, logger)
	if err != nil {
		return nil, nil, nil, err
	}

	ptcAddr := common.HexToAddress(cfg.Contract.PTCAddrHex)
	ptc, err := contract.NewPrivatixTokenContract(
		ptcAddr, conn.ethClient())
	if err != nil {
		return nil, nil, nil, err
	}

	pscAddr := common.HexToAddress(cfg.Contract.PSCAddrHex)
	psc, err := contract.NewPrivatixServiceContract(
		pscAddr, conn.ethClient())
	if err != nil {
		return nil, nil, nil, err
	}

	return conn, ptc, psc, nil
}

func (b *backendInstance) dropConnection() {
	// Close connections except currently in use.
	b.conn.closeIdleConnections()
	// Close connection currently in use.
	b.conn.close()
}

// addTimeout adds timeout to context.
func (b *backendInstance) addTimeout(
	ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx,
		time.Duration(b.cfg.Timeout)*second)
}

func (b *backendInstance) connectionControl() {
	timeout := time.Duration(b.cfg.CheckTimeout) * time.Second

	logger := b.logger.Add("method", "connectionControl",
		"timeout", timeout.String())

	connected := true

	for {
		<-time.After(timeout)

		ctx, _ := context.WithTimeout(context.Background(), timeout)
		_, err := b.conn.ethClient().HeaderByNumber(ctx, nil)
		if err != nil {
			connected = false
			logger.Warn("reconnecting to Ethereum")
			b.dropConnection()
			conn, ptc, psc, err := newInstance(b.cfg, b.logger)
			if err != nil {
				logger.Warn(fmt.Sprintf("failed to"+
					" reconnect to Ethereum: %s", err))
				continue
			}

			b.conn = conn
			b.psc = psc
			b.ptc = ptc
			continue
		}

		if !connected {
			connected = true
			logger.Warn("Ethereum communication restored")
			continue
		}

		logger.Debug("Ethereum communication checked")
	}
}

// LatestBlockNumber returns a block number from the current canonical chain.
func (b *backendInstance) LatestBlockNumber(ctx context.Context) (*big.Int,
	error) {
	ctx2, cancel := b.addTimeout(ctx)
	defer cancel()

	header, err := b.conn.ethClient().HeaderByNumber(ctx2, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get"+
			" latest block: %s", err)
	}
	return header.Number, err
}

// SuggestGasPrice retrieves the currently suggested gas price to allow a timely
// execution of a transaction.
func (b *backendInstance) SuggestGasPrice(
	ctx context.Context) (*big.Int, error) {
	ctx2, cancel := b.addTimeout(ctx)
	defer cancel()

	gasPrice, err := b.conn.ethClient().SuggestGasPrice(ctx2)
	if err != nil {
		return nil, fmt.Errorf("failed to get"+
			" suggested gas price: %s", err)
	}
	return gasPrice, err
}

// EstimateGas tries to estimate the gas needed to execute a specific
// transaction based on the current pending state of the backend blockchain.
func (b *backendInstance) EstimateGas(
	ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	ctx2, cancel := b.addTimeout(ctx)
	defer cancel()

	gas, err = b.conn.ethClient().EstimateGas(ctx2, call)
	if err != nil {
		return 0, fmt.Errorf("failed to estimated gas: %s", err)
	}
	return gas, err
}

// CooperativeClose calls cooperativeClose method of Privatix service contract.
func (b *backendInstance) CooperativeClose(opts *bind.TransactOpts,
	agent common.Address, block uint32, offeringHash [common.HashLength]byte,
	balance *big.Int, balanceSig, closingSig []byte) (*types.Transaction, error) {
	ctx2, cancel := b.addTimeout(opts.Context)
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
func (b *backendInstance) GetTransactionByHash(ctx context.Context,
	hash common.Hash) (*types.Transaction, bool, error) {
	ctx2, cancel := b.addTimeout(ctx)
	defer cancel()

	tx, pending, err := b.conn.ethClient().TransactionByHash(ctx2, hash)
	if err != nil {
		err = fmt.Errorf("failed to get transaction by hash: %s", err)
	}
	return tx, pending, err
}

// RegisterServiceOffering calls registerServiceOffering method of Privatix
// service contract.
func (b *backendInstance) RegisterServiceOffering(opts *bind.TransactOpts,
	offeringHash [common.HashLength]byte,
	minDeposit *big.Int, maxSupply uint16,
	somcType uint8, somcData data.Base64String) (*types.Transaction, error) {
	ctx2, cancel := b.addTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	tx, err := b.psc.RegisterServiceOffering(opts, offeringHash,
		minDeposit, maxSupply, somcType, string(somcData))
	if err != nil {
		return nil, fmt.Errorf(
			"failed to register service offering: %s", err)
	}
	return tx, nil
}

// PTCBalanceOf calls balanceOf method of Privatix token contract.
func (b *backendInstance) PTCBalanceOf(opts *bind.CallOpts,
	owner common.Address) (*big.Int, error) {
	ctx2, cancel := b.addTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	val, err := b.ptc.BalanceOf(opts, owner)
	if err != nil {
		err = fmt.Errorf("failed to get PTC balance: %s", err)
	}
	return val, err
}

// PTCIncreaseApproval calls increaseApproval method of Privatix token contract.
func (b *backendInstance) PTCIncreaseApproval(opts *bind.TransactOpts,
	spender common.Address, addedVal *big.Int) (*types.Transaction, error) {
	ctx2, cancel := b.addTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	tx, err := b.ptc.IncreaseApproval(opts, spender, addedVal)
	if err != nil {
		return nil, fmt.Errorf("failed to PTC increase approval: %s", err)
	}
	return tx, nil
}

// PSCBalanceOf calls balanceOf method of Privatix service contract.
func (b *backendInstance) PSCBalanceOf(opts *bind.CallOpts,
	owner common.Address) (*big.Int, error) {
	ctx2, cancel := b.addTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	val, err := b.psc.BalanceOf(opts, owner)
	if err != nil {
		err = fmt.Errorf("failed to get PSC balance: %s", err)
	}
	return val, err
}

// PSCAddBalanceERC20 calls addBalanceERC20 of Privatix service contract.
func (b *backendInstance) PSCAddBalanceERC20(opts *bind.TransactOpts,
	amount *big.Int) (*types.Transaction, error) {
	ctx2, cancel := b.addTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	tx, err := b.psc.AddBalanceERC20(opts, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to add ERC20 balance: %s", err)
	}
	return tx, nil
}

// PSCGetOfferingInfo calls getOfferingInfo of Privatix service contract.
func (b *backendInstance) PSCGetOfferingInfo(opts *bind.CallOpts,
	hash [common.HashLength]byte) (agentAddr common.Address,
	minDeposit *big.Int, maxSupply uint16, currentSupply uint16,
	updateBlockNumber uint32, active bool, err error) {
	ctx2, cancel := b.addTimeout(opts.Context)
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

// PSCGetChannelInfo calls getChannelInfo method of Privatix service contract.
func (b *backendInstance) PSCGetChannelInfo(opts *bind.CallOpts,
	client common.Address, agent common.Address,
	blockNumber uint32,
	hash [common.HashLength]byte) ([common.HashLength]byte,
	*big.Int, uint32, *big.Int, error) {
	ctx2, cancel := b.addTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2
	return b.psc.GetChannelInfo(opts, client, agent, blockNumber, hash)
}

// PSCCreateChannel calls createChannel method of Privatix service contract.
func (b *backendInstance) PSCCreateChannel(opts *bind.TransactOpts,
	agent common.Address, hash [common.HashLength]byte,
	deposit *big.Int) (*types.Transaction, error) {
	ctx2, cancel := b.addTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	tx, err := b.psc.CreateChannel(opts, agent, hash, deposit)
	if err != nil {
		err = fmt.Errorf("failed to create PSC channel: %s", err)
	}
	return tx, err
}

// PSCTopUpChannel calls topUpChannel method of Privatix service contract.
func (b *backendInstance) PSCTopUpChannel(opts *bind.TransactOpts,
	agent common.Address, blockNumber uint32, hash [common.HashLength]byte,
	deposit *big.Int) (*types.Transaction, error) {
	ctx2, cancel := b.addTimeout(opts.Context)
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
func (b *backendInstance) PSCUncooperativeClose(opts *bind.TransactOpts,
	agent common.Address, blockNumber uint32, hash [common.HashLength]byte,
	balance *big.Int) (*types.Transaction, error) {
	ctx2, cancel := b.addTimeout(opts.Context)
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
func (b *backendInstance) PSCReturnBalanceERC20(opts *bind.TransactOpts,
	amount *big.Int) (*types.Transaction, error) {
	ctx2, cancel := b.addTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	tx, err := b.psc.ReturnBalanceERC20(opts, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to return ERC20 balance: %s", err)
	}
	return tx, nil
}

// EthBalanceAt returns the wei balance of the given account.
func (b *backendInstance) EthBalanceAt(ctx context.Context,
	owner common.Address) (*big.Int, error) {
	ctx2, cancel := b.addTimeout(ctx)
	defer cancel()

	return b.conn.ethClient().BalanceAt(ctx2, owner, nil)
}

// PSCSettle calls settle method of Privatix service contract.
func (b *backendInstance) PSCSettle(opts *bind.TransactOpts,
	agent common.Address, blockNumber uint32,
	hash [common.HashLength]byte) (*types.Transaction, error) {
	ctx2, cancel := b.addTimeout(opts.Context)
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
func (b *backendInstance) PSCRemoveServiceOffering(opts *bind.TransactOpts,
	offeringHash [32]byte) (*types.Transaction, error) {
	ctx2, cancel := b.addTimeout(opts.Context)
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
func (b *backendInstance) PSCPopupServiceOffering(opts *bind.TransactOpts,
	offeringHash [32]byte, somcType uint8, somcData data.Base64String) (*types.Transaction, error) {
	ctx2, cancel := b.addTimeout(opts.Context)
	defer cancel()

	opts.Context = ctx2

	tx, err := b.psc.PopupServiceOffering(opts, offeringHash,
		somcType, string(somcData))
	if err != nil {
		err = fmt.Errorf("failed to pop up service offering: %v", err)
	}
	return tx, err
}

// FilterLogs executes a Ethereum filter query.
func (b *backendInstance) FilterLogs(ctx context.Context,
	q ethereum.FilterQuery) ([]types.Log, error) {
	return b.conn.ethClient().FilterLogs(ctx, q)
}

// HeaderByNumber returns a Ethereum block header from the current canonical
// chain. If number is nil, the latest known header is returned.
func (b *backendInstance) HeaderByNumber(ctx context.Context,
	number *big.Int) (*types.Header, error) {
	return b.conn.ethClient().HeaderByNumber(ctx, number)
}

// PTCAddress returns Privatix token contract address.
func (b *backendInstance) PTCAddress() common.Address {
	return common.HexToAddress(b.cfg.Contract.PTCAddrHex)
}

// PSCAddress returns Privatix service contract address.
func (b *backendInstance) PSCAddress() common.Address {
	return common.HexToAddress(b.cfg.Contract.PSCAddrHex)
}
