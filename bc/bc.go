package bc

import (
	"database/sql"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"gopkg.in/reform.v1"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/util/log"
)

// Config is a monitor configuration.
type Config struct {
	EthCallTimeout uint   // In milliseconds.
	InitialBlocks  uint64 // In Ethereum blocks.
	QueryPause     uint   // In milliseconds.
	RateAfter      uint   // Number of ethereum channel close events to initiate rating calculation after.
}

// NewConfig creates a default blockchain monitor configuration.
func NewConfig() *Config {
	return &Config{
		EthCallTimeout: 60000,
		InitialBlocks:  5760, // Is equivalent to 24 hours.
		QueryPause:     6000,
		RateAfter:      10,
	}
}

// NewMonitor creates blockchain monitor.
func NewMonitor(config *Config, client Client, queue Queue,
	db *reform.DB, logger log.Logger, pscAddr, ptcAddr common.Address, role string) (*Monitor, error) {

	if err := initLastProccessedBlock(logger, db); err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	if role == data.RoleClient {
		if err := initLastOfferingSearchFromBlock(logger, db); err != nil {
			logger.Error(err.Error())
			return nil, ErrInternal
		}
	}

	jm, err := newJobsMaker(db, logger, pscAddr, config.RateAfter, role)
	if err != nil {
		return nil, err
	}

	return &Monitor{
		client:         client,
		db:             db,
		logger:         logger.Add("type", "bc.Monitor"),
		Queue:          queue,
		RequestTimeout: time.Duration(config.EthCallTimeout) * time.Millisecond,
		RoundsInterval: time.Duration(config.QueryPause) * time.Millisecond,
		NextRound: func(latestBlock uint64) ([]ethereum.FilterQuery, func(*reform.TX) error, error) {
			if role == data.RoleAgent {
				return getAgentFilterQueries(logger, db, latestBlock, pscAddr, ptcAddr)
			}
			return getClientFilterQueries(logger, db, latestBlock, pscAddr, ptcAddr)
		},
		JobsForLog: jm.makeJobs,
	}, nil
}

func getAgentFilterQueries(logger log.Logger, db *reform.DB, latestBlock uint64,
	pscAddr, ptcAddr common.Address) ([]ethereum.FilterQuery, func(*reform.TX) error, error) {
	logger.Debug(fmt.Sprintf("getting agent filter queries on the latest block: %v", latestBlock))
	from, to, err := rangeOfInterest(db, latestBlock)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get range of interest: %v", err)
	}
	logger.Debug(fmt.Sprintf("range of interest from: %d, to: %d", from, to))
	if from >= to {
		return nil, nil, nil
	}
	addrs, err := getAddressesInUse(db)
	logger.Debug(fmt.Sprintf("addresses in use: %v", addrs))
	if err != nil {
		return nil, nil, fmt.Errorf("could not get addresses in use: %v", err)
	}
	if len(addrs) == 0 {
		return nil, nil, nil
	}
	return []ethereum.FilterQuery{
			{
				Addresses: []common.Address{pscAddr, ptcAddr},
				FromBlock: new(big.Int).SetUint64(from),
				ToBlock:   new(big.Int).SetUint64(to),
				Topics: [][]common.Hash{{
					eth.ServiceChannelCreated,
					eth.ServiceChannelToppedUp,
					eth.ServiceChannelCloseRequested,
					eth.ServiceOfferingCreated,
					eth.ServiceOfferingDeleted,
					eth.ServiceOfferingPopedUp,
					eth.ServiceCooperativeChannelClose,
					eth.ServiceUnCooperativeChannelClose,

					eth.TokenApproval,
					eth.TokenTransfer,
				}, addrs},
			},
			{
				Addresses: []common.Address{ptcAddr},
				FromBlock: new(big.Int).SetUint64(from),
				ToBlock:   new(big.Int).SetUint64(to),
				Topics: [][]common.Hash{{eth.TokenTransfer},
					nil, addrs},
			},
		}, func(tx *reform.TX) error {
			logger.Debug(fmt.Sprintf("updating last processed block to: %d", to))
			return updateLastProcessedBlock(tx.Querier, to)
		}, nil
}

func getClientFilterQueries(logger log.Logger, db *reform.DB, latestBlock uint64,
	pscAddr, ptcAddr common.Address) ([]ethereum.FilterQuery, func(*reform.TX) error, error) {
	logger.Debug(fmt.Sprintf("getting client filter queries on the latest block: %v", latestBlock))
	from, to, err := rangeOfInterest(db, latestBlock)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get range of interest: %v", err)
	}
	logger.Debug(fmt.Sprintf("range of interest from: %d, to: %v", from, to))
	addrs, err := getAddressesInUse(db)
	logger.Debug(fmt.Sprintf("addresses in use: %v", addrs))
	if err != nil {
		return nil, nil, fmt.Errorf("could not get addresses in use: %v", err)
	}
	if len(addrs) == 0 {
		return nil, nil, nil
	}
	ret := make([]ethereum.FilterQuery, 0)
	if from < to {
		ret = append(ret, ethereum.FilterQuery{
			Addresses: []common.Address{pscAddr, ptcAddr},
			FromBlock: new(big.Int).SetUint64(from),
			ToBlock:   new(big.Int).SetUint64(to),
			Topics: [][]common.Hash{
				{
					eth.TokenTransfer,
					eth.ServiceChannelToppedUp,
					eth.ServiceChannelCloseRequested,
				},
				nil,
				addrs},
		}, ethereum.FilterQuery{
			Addresses: []common.Address{ptcAddr},
			FromBlock: new(big.Int).SetUint64(from),
			ToBlock:   new(big.Int).SetUint64(to),
			Topics: [][]common.Hash{
				{
					eth.TokenTransfer, eth.TokenApproval,
				},
				addrs,
			},
		}, ethereum.FilterQuery{
			Addresses: []common.Address{pscAddr},
			FromBlock: new(big.Int).SetUint64(from),
			ToBlock:   new(big.Int).SetUint64(to),
			Topics: [][]common.Hash{
				{
					eth.ServiceChannelCreated,
					eth.ServiceOfferingCreated,
					eth.ServiceOfferingDeleted,
					eth.ServiceOfferingPopedUp,
					eth.ServiceCooperativeChannelClose,
					eth.ServiceUnCooperativeChannelClose,
				},
			},
		})
	}
	// Get up block for backward offerings search.
	upBlock := latestBlock
	var startBlock data.Setting
	if err := db.FindOneTo(&startBlock, "key", data.SettingClientMonitoringStartBlock); err == sql.ErrNoRows {
		logger.Debug("recording monitoring start block to " + fmt.Sprint(upBlock))
		startBlock.Key = data.SettingClientMonitoringStartBlock
		startBlock.Value = fmt.Sprint(upBlock)
		startBlock.Permissions = data.AccessDenied
		tmp := "Block from which (Client) monitoring started started."
		startBlock.Description = &tmp
		startBlock.Name = "Client monitoring start block"
		if err := db.Insert(&startBlock); err != nil {
			return nil, nil, err
		}
	} else if err == nil {
		val, err := strconv.ParseUint(startBlock.Value, 10, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("could not parse `%s`: %v", data.SettingClientMonitoringStartBlock, err)
		}
		upBlock = val
	} else if err != nil {
		return nil, nil, fmt.Errorf("could not get `%s`: %v", data.SettingClientMonitoringStartBlock, err)
	}
	offeringsFrom, offeringsTo, err := offeringsRangeOfInterest(db, upBlock)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get offerings range of interest: %v", err)
	}
	logger.Debug(fmt.Sprintf("offerings range of interest from: %v, to: %v", offeringsFrom, offeringsTo))
	if offeringsFrom < offeringsTo {
		ret = append(ret, ethereum.FilterQuery{
			Addresses: []common.Address{pscAddr},
			FromBlock: new(big.Int).SetUint64(offeringsFrom),
			ToBlock:   new(big.Int).SetUint64(offeringsTo),
			Topics: [][]common.Hash{
				{
					eth.ServiceOfferingCreated,
					eth.ServiceOfferingPopedUp,
				},
			},
		})
	}
	if len(ret) == 0 {
		return nil, nil, nil
	}
	return ret, func(tx *reform.TX) error {
		logger.Debug(fmt.Sprintf("updating last processed block to: %d", to))
		if err := updateLastProcessedBlock(tx.Querier, to); err != nil {
			return err
		}
		logger.Debug(fmt.Sprintf("updating last processed blocks in backward offerings search to=%d", offeringsFrom))
		return data.UpdateUint64Setting(tx.Querier, data.SettingLastBackSearchBlock, offeringsFrom)
	}, nil
}

func getAddressesInUse(db *reform.DB) ([]common.Hash, error) {
	rows, err := db.Query(`SELECT eth_addr FROM accounts WHERE in_use`)
	if err != nil {
		return nil, fmt.Errorf("could not query account addresses: %v", err)
	}
	defer rows.Close()

	var addresses []common.Hash
	for rows.Next() {
		var addrHex string
		if err := rows.Scan(&addrHex); err != nil {
			return nil, fmt.Errorf("could not read account address: %v", err)
		}
		addresses = append(addresses, common.HexToHash(addrHex))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("could not read account addresses: %v", err)
	}
	return addresses, nil
}

func initLastProccessedBlock(logger log.Logger, db *reform.DB) error {
	logger = logger.Add("method", "initLastProcessedBlock")
	_, err := db.FindOneFrom(data.SettingTable, "key", data.SettingLastProcessedBlock)
	if err == sql.ErrNoRows {
		logger.Debug("last processed block not set")
		desc := "Last block number in blockchain stores last proccessed block."
		return db.Insert(&data.Setting{
			Key:         data.SettingLastProcessedBlock,
			Value:       "0",
			Permissions: data.ReadOnly,
			Description: &desc,
			Name:        "last processed block",
		})
	}
	logger.Debug("last processed block already exists")
	return nil
}

func initLastOfferingSearchFromBlock(logger log.Logger, db *reform.DB) error {
	logger = logger.Add("method", "initLastOfferingSearchFromBlock")
	_, err := db.FindOneFrom(data.SettingTable, "key", data.SettingLastBackSearchBlock)
	if err == sql.ErrNoRows {
		logger.Debug("offerings last processed block not set")
		desc := "'On client, the last block offerings searched from."
		return db.Insert(&data.Setting{
			Key:         data.SettingLastBackSearchBlock,
			Value:       "0",
			Permissions: data.ReadOnly,
			Description: &desc,
			Name:        "last back search block",
		})
	}
	logger.Debug("offerings last processed block already exists")
	return nil
}

func updateLastProcessedBlock(db *reform.Querier, to uint64) error {
	return data.UpdateUint64Setting(db, data.SettingLastProcessedBlock, to)
}
