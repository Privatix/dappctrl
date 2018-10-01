package worker

import (
	"github.com/privatix/dappctrl/util/errors"
)

// Errors returned by workers.
const (
	ErrInternal errors.Error = 0x25E9<<8 + iota
	ErrInvalidJob
	ErrAccountNotFound
	ErrChannelNotFound
	ErrOfferingNotFound
	ErrUserNotFound
	ErrEndpointNotFound
	ErrTemplateNotFound
	ErrTemplateByHashNotFound
	ErrEthLogNotFound
	ErrProductNotFound
	ErrParseJobData
	ErrParsePrivateKey
	ErrParseEthAddr
	ErrParseOfferingHash
	ErrParseEthLog
	ErrWrongLogNonIndexedArgsNumber
	ErrWrongLogTopicsNumber
	ErrInsufficientPTCBalance
	ErrInsufficientPSCBalance
	ErrInsufficientEthBalance
	ErrSmallDeposit
	ErrFindApprovalBalanceData
	ErrPTCRetrieveBalance
	ErrPTCIncreaseApproval
	ErrPSCRetrieveBalance
	ErrPSCAddBalance
	ErrPSCReturnBalance
	ErrPSCCooperativeClose
	ErrPSCRegisterOffering
	ErrPSCRemoveOffering
	ErrPSCPopUpOffering
	ErrPSCOfferingSupply
	ErrPSCCreateChannel
	ErrPSCGetChannelInfo
	ErrPSCSettle
	ErrPSCTopUpChannel
	ErrPSCUncooperativeClose
	ErrEthGetTransaction
	ErrEthTxInPendingState
	ErrEthRetrieveBalance
	ErrEthLogChannelMismatch
	ErrEthLatestBlockNumber
	ErrRecoverClientPubKey
	ErrTerminateChannel
	ErrAddJob
	ErrSignClosingMsg
	ErrNoReceiptSignature
	ErrMakeEndpointMsg
	ErrEndpointMsgSeal
	ErrGeneratePasswordHash
	ErrPublishEndpoint
	ErrPublishOffering
	ErrFindOfferings
	ErrSOMCNoOfferings
	ErrOfferNotRegistered
	ErrOfferNotCorrespondToTemplate
	ErrChannelReceiptBalance
	ErrInvalidChannelStatus
	ErrInvalidServiceStatus
	ErrOfferingNoSupply
	ErrGetEndpoint
	ErrDecryptEndpointMsg
	ErrInvalidEndpoint
	ErrFailedStopService
	ErrFailedStartService
	ErrChallengePeriodIsNotOver
	ErrWrongOfferingMsgSignature
	ErrOfferingExists
	ErrUncompletedJobsExists
	ErrOfferingNotActive
	ErrPopUpOfferingTryAgain
	ErrOfferingDeposit
)

var errMsgs = errors.Messages{
	ErrInternal:                     "internal server error",
	ErrInvalidJob:                   "unexpected job or job related type",
	ErrAccountNotFound:              "could not find account record",
	ErrChannelNotFound:              "could not find channel record",
	ErrOfferingNotFound:             "could not find offering record",
	ErrUserNotFound:                 "could not find user(client) record",
	ErrEndpointNotFound:             "could not find endpoint record",
	ErrTemplateNotFound:             "could not find template record",
	ErrTemplateByHashNotFound:       "could not find template by given hash",
	ErrEthLogNotFound:               "could not find eth log",
	ErrProductNotFound:              "could not find a product",
	ErrParseJobData:                 "unable to parse job's data",
	ErrParseEthAddr:                 "unable to parse ethereum address",
	ErrParseOfferingHash:            "unable to parse offering hash",
	ErrParseEthLog:                  "unable to parse ethereum log",
	ErrParsePrivateKey:              "unable to parse account's private key",
	ErrWrongLogNonIndexedArgsNumber: "wrong eth log's non-indexed arguments",
	ErrWrongLogTopicsNumber:         "wrong number of ethereum log topics",
	ErrInsufficientPTCBalance:       "insufficient PTC balance",
	ErrInsufficientPSCBalance:       "insufficient PSC balance",
	ErrInsufficientEthBalance:       "insufficient eth balance",
	ErrSmallDeposit:                 "deposit is too small",
	ErrFindApprovalBalanceData:      "could not find PTC approval data",
	ErrPTCRetrieveBalance:           "could not get PTC balance",
	ErrPTCIncreaseApproval:          "could not increase approval (PTC)",
	ErrPSCRetrieveBalance:           "could not get PSC balance",
	ErrPSCAddBalance:                "could not add balance (PSC)",
	ErrPSCReturnBalance:             "could not return balance (PSC)",
	ErrPSCCooperativeClose:          "could not cooperative close (PSC)",
	ErrPSCRegisterOffering:          "could not register an offering (PSC)",
	ErrPSCRemoveOffering:            "could not remove an offering (PSC)",
	ErrPSCPopUpOffering:             "could not pop up an offering (PSC)",
	ErrPSCOfferingSupply:            "could not get offering supply (PSC)",
	ErrPSCCreateChannel:             "could not create a channel (PSC)",
	ErrPSCGetChannelInfo:            "could not get a channel info (PSC)",
	ErrPSCSettle:                    "could not settle a channel (PSC)",
	ErrPSCTopUpChannel:              "could not top up a channel (PSC)",
	ErrPSCUncooperativeClose:        "could not uncooperatively close",
	ErrEthGetTransaction:            "could not get ethereum transaction",
	ErrEthTxInPendingState:          "pending state of a transaction",
	ErrEthRetrieveBalance:           "could not get ethereum balance",
	ErrEthLogChannelMismatch:        "channel does not correspond to eth log",
	ErrEthLatestBlockNumber:         "could not get the latest block number",
	ErrRecoverClientPubKey:          "could not recover client's pub key",
	ErrTerminateChannel:             "could not terminte a channel",
	ErrAddJob:                       "failed to add a job",
	ErrSignClosingMsg:               "could not sign closing msg",
	ErrNoReceiptSignature:           "no receipt signature in channel",
	ErrMakeEndpointMsg:              "could not make endpoint message",
	ErrEndpointMsgSeal:              "could not seal endpoint message",
	ErrGeneratePasswordHash:         "failed to generate password hash",
	ErrPublishEndpoint:              "could not publish endpoint message",
	ErrPublishOffering:              "could not publish an offering",
	ErrFindOfferings:                "could not find offerings",
	ErrSOMCNoOfferings:              "no offering returned from somc",
	ErrOfferNotRegistered:           "offering is not registered",
	ErrOfferNotCorrespondToTemplate: "offer not correspond to template",
	ErrChannelReceiptBalance:        "receipt balance more than deposit",
	ErrInvalidChannelStatus:         "channel status forbids procedure",
	ErrInvalidServiceStatus:         "service status forbids procedure",
	ErrOfferingNoSupply:             "no supply",
	ErrGetEndpoint:                  "could not get endpoint from SOMC",
	ErrDecryptEndpointMsg:           "could not decrypt endpoint message",
	ErrInvalidEndpoint:              "endpoint doesn't correspond to schema",
	ErrFailedStopService:            "failed to stop a service",
	ErrFailedStartService:           "failed to start a service",
	ErrChallengePeriodIsNotOver:     "challenge period is not over",
	ErrWrongOfferingMsgSignature:    "wrong offering msg's signature",
	ErrOfferingExists:               "offering with a given hash exists",
	ErrUncompletedJobsExists:        "active offering related jobs exists",
	ErrOfferingNotActive:            "offering is inactive",
	ErrPopUpOfferingTryAgain:        "could not pop up, try again later",
	ErrOfferingDeposit:              "incorrect offering deposit",
}

func init() {
	errors.InjectMessages(errMsgs)
}
