package prepare

import (
	"github.com/privatix/dappctrl/util/errors"
)

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/svc/dappvpn/prepare") = 0x23BD
	ErrMakeConfig errors.Error = 0x23BD<<8 + iota
)

var errMsgs = errors.Messages{
	ErrMakeConfig: "failed to make client configuration files",
}

func init() { errors.InjectMessages(errMsgs) }
