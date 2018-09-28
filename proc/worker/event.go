package worker

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
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

type logOfferingPopUpInput struct {
	agentAddr    common.Address
	offeringHash common.Hash
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

func extractLogChannelToppedUp(
	logger log.Logger, log *data.JobEthLog) (*logChannelTopUpInput, error) {
	dataUnpacked, err := logChannelTopUpDataArguments.UnpackValues(log.Data)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrParseEthLog
	}

	if len(dataUnpacked) != 2 {
		return nil, ErrWrongLogNonIndexedArgsNumber
	}

	openBlockNum, ok := dataUnpacked[0].(uint32)
	if !ok {
		return nil, ErrParseEthLog
	}

	addedDeposit, ok := dataUnpacked[1].(*big.Int)
	if !ok {
		return nil, ErrParseEthLog
	}

	if len(log.Topics) != 4 {
		return nil, ErrWrongLogTopicsNumber
	}

	agentAddr := common.BytesToAddress(log.Topics[1].Bytes())
	clientAddr := common.BytesToAddress(log.Topics[2].Bytes())
	offeringHash := log.Topics[3]

	return &logChannelTopUpInput{
		agentAddr:    agentAddr,
		clientAddr:   clientAddr,
		offeringHash: offeringHash,
		openBlockNum: openBlockNum,
		addedDeposit: addedDeposit,
	}, nil
}

func extractLogChannelCreated(logger log.Logger,
	log *data.JobEthLog) (*logChannelCreatedInput, error) {
	dataUnpacked, err := logChannelCreatedDataArguments.UnpackValues(log.Data)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrParseEthLog
	}

	if len(dataUnpacked) != 2 {
		return nil, ErrWrongLogNonIndexedArgsNumber
	}

	deposit, ok := dataUnpacked[0].(*big.Int)
	if !ok {
		return nil, ErrParseEthLog
	}

	authHashB, ok := dataUnpacked[1].([common.HashLength]byte)
	if !ok {
		return nil, ErrParseEthLog
	}

	if len(log.Topics) != 4 {
		return nil, ErrWrongLogTopicsNumber
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

func extractLogOfferingCreated(logger log.Logger,
	log *data.JobEthLog) (*logOfferingCreatedInput, error) {
	if len(log.Topics) != 4 {
		return nil, ErrWrongLogTopicsNumber
	}

	dataUnpacked, err := logOfferingCreatedDataArguments.UnpackValues(log.Data)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrParseJobData
	}

	if len(dataUnpacked) != 1 {
		return nil, ErrWrongLogNonIndexedArgsNumber
	}

	curSupply, ok := dataUnpacked[0].(uint16)
	if !ok {
		return nil, ErrParseJobData
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

func extractLogOfferingPopUp(logger log.Logger,
	log *data.JobEthLog) (*logOfferingPopUpInput, error) {
	if len(log.Topics) != 3 {
		return nil, ErrWrongLogTopicsNumber
	}

	agentAddr := common.BytesToAddress(log.Topics[1].Bytes())
	offeringHash := log.Topics[2]

	return &logOfferingPopUpInput{
		agentAddr:    agentAddr,
		offeringHash: offeringHash,
	}, nil
}
