package monitor

import (
	"math/big"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"

	"github.com/privatix/dappctrl/eth"
)

func (m *Monitor) agentQueries(from, to uint64,
	psc, ptc common.Address) ([]ethereum.FilterQuery, error) {
	addresses, err := m.getAddressesInUse()
	if err != nil {
		return nil, err
	}
	return []ethereum.FilterQuery{
		{
			Addresses: []common.Address{psc, ptc},
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
			}, addresses},
		},
		{
			Addresses: []common.Address{ptc},
			FromBlock: new(big.Int).SetUint64(from),
			ToBlock:   new(big.Int).SetUint64(to),
			Topics: [][]common.Hash{{eth.TokenTransfer},
				nil, addresses},
		},
	}, nil
}

func (m *Monitor) clientQueries(from, to uint64,
	psc, ptc common.Address) ([]ethereum.FilterQuery, error) {
	addresses, err := m.getAddressesInUse()
	if err != nil {
		return nil, err
	}
	return []ethereum.FilterQuery{
		{
			Addresses: []common.Address{psc, ptc},
			FromBlock: new(big.Int).SetUint64(from),
			ToBlock:   new(big.Int).SetUint64(to),
			Topics: [][]common.Hash{
				{
					eth.TokenTransfer,
					eth.ServiceChannelToppedUp,
					eth.ServiceChannelCloseRequested,
				},
				nil,
				addresses},
		},
		{
			Addresses: []common.Address{ptc},
			FromBlock: new(big.Int).SetUint64(from),
			ToBlock:   new(big.Int).SetUint64(to),
			Topics: [][]common.Hash{
				{
					eth.TokenTransfer, eth.TokenApproval,
				},
				addresses,
			},
		},
		{
			Addresses: []common.Address{psc},
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
		},
	}, nil
}

func (m *Monitor) getAddressesInUse() ([]common.Hash, error) {
	logger := m.logger.Add("method", "getAddressesInUse")

	rows, err := m.db.Query(`SELECT eth_addr
		                   FROM accounts
		                  WHERE in_use`)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrFailedToGetActiveAccounts
	}
	defer rows.Close()

	var addresses []common.Hash
	for rows.Next() {
		var addrHex string
		if err := rows.Scan(&addrHex); err != nil {
			logger.Error(err.Error())
			return nil, ErrFailedToScanRows
		}
		addresses = append(addresses, common.HexToHash(addrHex))
	}
	if err := rows.Err(); err != nil {
		m.logger.Error(err.Error())
		return nil, ErrFailedToTraverseAddresses
	}
	return addresses, nil
}
