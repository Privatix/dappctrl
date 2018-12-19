package sess

import (
	"github.com/privatix/dappctrl/util/errors"
)

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/sess") = 0x12DD
	ErrAccessDenied errors.Error = 0x12DD<<8 + iota
	ErrChannelNotFound
	ErrBadClientPassword
	ErrInternal
	ErrNonActiveChannel
	ErrSessionNotFound
	ErrEndpointNotFound
	ErrBadProductConfig
)

var errMsgs = errors.Messages{
	ErrAccessDenied:      "access denied",
	ErrChannelNotFound:   "channel not found",
	ErrBadClientPassword: "bad client password",
	ErrInternal:          "internal server error",
	ErrNonActiveChannel:  "non-active channel",
	ErrSessionNotFound:   "session not found",
	ErrEndpointNotFound:  "endpoint not found",
	ErrBadProductConfig:  "bad product config",
}

func init() {
	errors.InjectMessages(errMsgs)
}
