package billing

import (
	"github.com/privatix/dappctrl/util/errors"
)

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/agent/bill") = 0x6D62
	ErrInput errors.Error = 0x6D62<<8 + iota
)

var errMsgs = errors.Messages{
	ErrInput: "one or more input parameters is wrong",
}

func init() { errors.InjectMessages(errMsgs) }
