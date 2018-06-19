package worker

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/privatix/dappctrl/data"
)

type logChannelTopUpInput struct {
	agentAddr    common.Address
	clientAddr   common.Address
	offeringHash common.Hash
	openBlockNum uint32
	addedDeposit *big.Int
}

type logChannelCreatedInput struct {
	agentAddr          common.Address
	clientAddr         common.Address
	offeringHash       common.Hash
	deposit            *big.Int
	authenticationHash common.Hash
}

type logOfferingCreatedInput struct {
	agentAddr    common.Address
	offeringHash common.Hash
	minDeposit   *big.Int
	maxSupply    uint16
}

var (
	logChannelTopUpDataArguments    abi.Arguments
	logChannelCreatedDataArguments  abi.Arguments
	logOfferingCreatedDataArguments abi.Arguments
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

	abiBytes32, err := abi.NewType("bytes32")
	if err != nil {
		panic(err)
	}

	abiUint16, err := abi.NewType("uint16")
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

	logChannelCreatedDataArguments = abi.Arguments{
		{
			Type: abiUint192,
		},
		{
			Type: abiBytes32,
		},
	}

	logOfferingCreatedDataArguments = abi.Arguments{
		{
			Type: abiUint16,
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

	if len(log.Topics) != 3 {
		return nil, fmt.Errorf("wrong number of topics")
	}

	agentAddr := common.BytesToAddress(log.Topics[0].Bytes())
	clientAddr := common.BytesToAddress(log.Topics[1].Bytes())
	offeringHash := log.Topics[2]

	return &logChannelTopUpInput{
		agentAddr:    agentAddr,
		clientAddr:   clientAddr,
		offeringHash: offeringHash,
		openBlockNum: openBlockNum,
		addedDeposit: addedDeposit,
	}, nil
}

func extractLogChannelCreated(log *data.EthLog) (*logChannelCreatedInput, error) {
	dataBytes, err := data.ToBytes(log.Data)
	if err != nil {
		return nil, fmt.Errorf("could not decode log data: %v", err)
	}

	dataUnpacked, err := logChannelCreatedDataArguments.UnpackValues(dataBytes)
	if err != nil {
		return nil, fmt.Errorf("could not unpack using %T: %v",
			logChannelCreatedDataArguments, err)
	}

	if len(dataUnpacked) != 2 {
		return nil, fmt.Errorf("wrong number of non-indexed arguments")
	}

	deposit, ok := dataUnpacked[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("could not parse deposit")
	}

	authHashB, ok := dataUnpacked[1].([common.HashLength]byte)
	if !ok {
		return nil, fmt.Errorf("could not parse authentication hash")
	}

	if len(log.Topics) != 4 {
		return nil, fmt.Errorf(
			"wrong number of topics, wanted: %v, got: %v",
			4, len(log.Topics))
	}

	agentAddr := common.BytesToAddress(log.Topics[1].Bytes())
	clientAddr := common.BytesToAddress(log.Topics[2].Bytes())
	offeringHash := log.Topics[3]

	return &logChannelCreatedInput{
		agentAddr:          agentAddr,
		clientAddr:         clientAddr,
		offeringHash:       offeringHash,
		deposit:            deposit,
		authenticationHash: common.Hash(authHashB),
	}, nil
}

func extractLogOfferingCreated(log *data.EthLog) (*logOfferingCreatedInput, error) {
	if len(log.Topics) != 4 {
		return nil, fmt.Errorf(
			"wrong number of topics, wanted: %v, got: %v",
			4, len(log.Topics))
	}

	dataBytes, err := data.ToBytes(log.Data)
	if err != nil {
		return nil, err
	}

	dataUnpacked, err := logOfferingCreatedDataArguments.UnpackValues(dataBytes)
	if err != nil {
		return nil, err
	}

	if len(dataUnpacked) != 1 {
		return nil, fmt.Errorf("wrong number of non-indexed arguments")
	}

	curSupply, ok := dataUnpacked[0].(uint16)
	if !ok {
		return nil, fmt.Errorf("could not decode event data")
	}

	agentAddr := common.BytesToAddress(log.Topics[1].Bytes())
	offeringHash := log.Topics[2]
	minDeposit := big.NewInt(0).SetBytes(log.Topics[3].Bytes())

	return &logOfferingCreatedInput{
		agentAddr:    agentAddr,
		offeringHash: offeringHash,
		minDeposit:   minDeposit,
		maxSupply:    curSupply,
	}, nil
}
