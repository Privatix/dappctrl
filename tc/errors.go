package tc

import "github.com/privatix/dappctrl/util/errors"

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/tc") = 0xB8DC
	ErrBadClientIP errors.Error = 0xB8DC<<8 + iota
)

var errMsgs = errors.Messages{
	ErrBadClientIP: "bad client IP",
}

func init() { errors.InjectMessages(errMsgs) }
