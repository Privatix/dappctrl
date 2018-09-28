package monitor

import "github.com/privatix/dappctrl/util/errors"

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/monitor") = 0xE464
	ErrFailedToFetchLogs errors.Error = 0xE464<<8 + iota
	ErrFailedToGetActiveAccounts
	ErrFailedToScanRows
	ErrFailedToTraverseAddresses
	ErrFailedToGetHeaderByNumber
	ErrFailedToParseABI
	ErrFailedToUnpack
	ErrWrongNumberOfEventArgs
	ErrWrongBlockArgumentType
	ErrUnsupportedTopic
)

var errMsgs = errors.Messages{
	ErrFailedToFetchLogs:         "failed to fetch logs from blockchain",
	ErrFailedToGetActiveAccounts: "failed to get active accounts from db",
	ErrFailedToScanRows:          "failed to scan rows",
	ErrFailedToTraverseAddresses: "failed to traverse the selected addresses",
	ErrFailedToGetHeaderByNumber: "failed to get header by block number" +
		" from blockchain",
	ErrFailedToParseABI:       "failed to parse ABI from string",
	ErrFailedToUnpack:         "failed to unpack arguments from event",
	ErrWrongNumberOfEventArgs: "wrong number of event arguments",
	ErrWrongBlockArgumentType: "wrong block number argument type",
	ErrUnsupportedTopic:       "unsupported topic",
}

func init() { errors.InjectMessages(errMsgs) }
