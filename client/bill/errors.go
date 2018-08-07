package bill

import (
	"github.com/privatix/dappctrl/util/errors"
)

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/client/bill") = 0xEB0A
	ErrAlreadyRunning errors.Error = 0xEB0A<<8 + iota
	ErrMonitorClosed
	ErrGetConsumedUnits
	ErrGetOffering
	ErrUpdateReceiptBalance
)

var errMsgs = errors.Messages{
	ErrAlreadyRunning:       "already running",
	ErrMonitorClosed:        "client billing monitor closed",
	ErrGetConsumedUnits:     "failed to get consumed units",
	ErrGetOffering:          "failed to get offering",
	ErrUpdateReceiptBalance: "failed to update receipt balance",
}

func init() { errors.InjectMessages(errMsgs) }
