package prepare

import (
	"github.com/privatix/dappctrl/util/errors"
)

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/svc/dappvpn/prepare") = 0x23BD
	ErrGetEndpoint errors.Error = 0x23BD<<8 + iota
	ErrMakeConfig
)

var errMsgs = errors.Messages{
	ErrGetEndpoint: "failed to get endpoint",
	ErrMakeConfig:  "failed to make client configuration files",
}

func init() { errors.InjectMessages(errMsgs) }
