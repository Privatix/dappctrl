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
	agentAddr    common.Address
	clientAddr   common.Address
	offeringHash common.Hash
	deposit      *big.Int
}

type logOfferingCreatedInput struct {
	agentAddr     common.Address
	offeringHash  common.Hash
	minDeposit    *big.Int
	currentSupply uint16
	somcType      uint8
	somcData      data.Base64String
}

type logOfferingPopUpInput struct {
	agentAddr     common.Address
	offeringHash  common.Hash
	minDeposit    *big.Int
	currentSupply uint16
	somcType      uint8
	somcData      data.Base64String
}

var (
	logChannelTopUpDataArguments    abi.Arguments
	logChannelCreatedDataArguments  abi.Arguments
	logOfferingCreatedDataArguments abi.Arguments
	logOfferingPopUpDataArguments   abi.Arguments
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

	abiUint16, err := abi.NewType("uint16")
	if err != nil {
		panic(err)
	}

	abiUint8, err := abi.NewType("uint8")
	if err != nil {
		panic(err)
	}

	abiString, err := abi.NewType("string")
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
	}

	logOfferingCreatedDataArguments = abi.Arguments{
		{
			Type: abiUint16,
		},
		{
			Type: abiUint8,
		},
		{
			Type: abiString,
		},
	}

	logOfferingPopUpDataArguments = abi.Arguments{
		{
			Type: abiUint16,
		},
		{
			Type: abiUint8,
		},
		{
			Type: abiString,
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

	if len(dataUnpacked) != 1 {
		return nil, ErrWrongLogNonIndexedArgsNumber
	}

	deposit, ok := dataUnpacked[0].(*big.Int)
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
		agentAddr:    agentAddr,
		clientAddr:   clientAddr,
		offeringHash: offeringHash,
		deposit:      deposit,
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

	if len(dataUnpacked) != 3 {
		return nil, ErrWrongLogNonIndexedArgsNumber
	}

	curSupply, ok := dataUnpacked[0].(uint16)
	if !ok {
		return nil, ErrParseJobData
	}

	somcType, ok := dataUnpacked[1].(uint8)
	if !ok {
		return nil, ErrParseJobData
	}

	somcData, ok := dataUnpacked[2].(string)
	if !ok {
		return nil, ErrParseJobData
	}

	agentAddr := common.BytesToAddress(log.Topics[1].Bytes())
	offeringHash := log.Topics[2]
	minDeposit := big.NewInt(0).SetBytes(log.Topics[3].Bytes())

	return &logOfferingCreatedInput{
		agentAddr:     agentAddr,
		offeringHash:  offeringHash,
		minDeposit:    minDeposit,
		currentSupply: curSupply,
		somcType:      somcType,
		somcData:      data.Base64String(somcData),
	}, nil
}

func extractLogOfferingPopUp(logger log.Logger,
	log *data.JobEthLog) (*logOfferingPopUpInput, error) {
	if len(log.Topics) != 4 {
		return nil, ErrWrongLogTopicsNumber
	}

	dataUnpacked, err := logOfferingCreatedDataArguments.UnpackValues(log.Data)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrParseJobData
	}

	if len(dataUnpacked) != 3 {
		return nil, ErrWrongLogNonIndexedArgsNumber
	}

	currentSupply, ok := dataUnpacked[0].(uint16)
	if !ok {
		return nil, ErrParseJobData
	}

	somcType, ok := dataUnpacked[1].(uint8)
	if !ok {
		return nil, ErrParseJobData
	}

	somcData, ok := dataUnpacked[2].(string)
	if !ok {
		return nil, ErrParseJobData
	}

	agentAddr := common.BytesToAddress(log.Topics[1].Bytes())
	offeringHash := log.Topics[2]
	minDeposit := new(big.Int).SetBytes(log.Topics[3].Bytes())

	return &logOfferingPopUpInput{
		agentAddr:     agentAddr,
		offeringHash:  offeringHash,
		minDeposit:    minDeposit,
		currentSupply: currentSupply,
		somcType:      somcType,
		somcData:      data.Base64String(somcData),
	}, nil
}
