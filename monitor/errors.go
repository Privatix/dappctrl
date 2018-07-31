package monitor

import (
	"github.com/privatix/dappctrl/util/errors"
)

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/monitor") = 0xE464
	ErrGetRangeOfInterest errors.Error = 0xE464<<8 + iota
	ErrGetAddressesInUse
	ErrFetchLogs
	ErrInsertLogEvent
	ErrGetActiveAccounts
	ErrScanRows
	ErrTraverseAddresses
	ErrGetHeaderByNumber
	ErrInput
	ErrParseABI
	ErrSelectLogsEntries
	ErrFetchLogsFromDB
	ErrUnpack
	ErrNumberOfEventArgs
	ErrBlockArgumentType
	ErrUnsupportedTopic
)

var errMsgs = errors.Messages{
	ErrGetRangeOfInterest: "failed to get range of interest",
	ErrGetAddressesInUse:  "failed to get addresses in use",
	ErrFetchLogs:          "could not fetch logs from blockchain",
	ErrInsertLogEvent:     "failed to insert a log event into database",
	ErrGetActiveAccounts:  "failed to query active accounts from database",
	ErrScanRows:           "failed to scan rows",
	ErrTraverseAddresses:  "failed to traverse the selected addresses",
	ErrGetHeaderByNumber: "failed to get header by block number" +
		" from blockchain",
	ErrInput:             "one or more input parameters is wrong",
	ErrParseABI:          "failed to parse ABI from string",
	ErrSelectLogsEntries: "failed to select Ethereum log entries",
	ErrFetchLogsFromDB:   "could not fetch logs from database",
	ErrUnpack:            "failed to unpack arguments from event",
	ErrNumberOfEventArgs: "wrong number of event arguments",
	ErrBlockArgumentType: "wrong block number argument type",
	ErrUnsupportedTopic:  "unsupported topic",
}

func init() { errors.InjectMessages(errMsgs) }
