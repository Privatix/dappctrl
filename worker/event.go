package worker

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/privatix/dappctrl/data"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type logChannelTopUpInput struct {
	agentAddr    common.Address
	clientAddr   common.Address
	offeringHash common.Hash
	openBlockNum uint32
	addedDeposit *big.Int
}

var (
	logChannelTopUpDataArguments abi.Arguments
)

func init() {
	abiUint32, err := abi.NewType("uint32")
	if err != nil {
		panic(err)
	}

	abiUint192, err := abi.NewType("uint192")
	if err != nil {
		panic(err)
	}

	logChannelTopUpDataArguments = abi.Arguments{
		{
			Type: abiUint32,
		},
		{
			Type: abiUint192,
		},
	}
}

func extractLogChannelToppedUp(log *data.EthLog) (*logChannelTopUpInput, error) {
	dataBytes, err := data.ToBytes(log.Data)
	if err != nil {
		return nil, err
	}

	dataUnpacked, err := logChannelTopUpDataArguments.UnpackValues(dataBytes)
	if err != nil {
		return nil, err
	}

	if len(dataUnpacked) != 2 {
		return nil, fmt.Errorf("wrong number of non-indexed arguments")
	}

	openBlockNum, ok := dataUnpacked[0].(uint32)
	if !ok {
		return nil, fmt.Errorf("could not decode event data")
	}

	addedDeposit, ok := dataUnpacked[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("could not decode event data")
	}

	topics := []common.Hash{}

	if err = json.Unmarshal(log.Topics, &topics); err != nil {
		return nil, fmt.Errorf("could not decode event topics: %v", err)
	}

	if len(topics) != 3 {
		return nil, fmt.Errorf("wrong number of topics")
	}

	agentAddr := common.BytesToAddress(topics[0].Bytes())
	clientAddr := common.BytesToAddress(topics[1].Bytes())
	offeringHash := topics[2]

	return &logChannelTopUpInput{
		agentAddr:    agentAddr,
		clientAddr:   clientAddr,
		offeringHash: offeringHash,
		openBlockNum: openBlockNum,
		addedDeposit: addedDeposit,
	}, nil
}
