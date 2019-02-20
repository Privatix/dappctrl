package nat

import "github.com/privatix/dappctrl/util/errors"

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/nat") = 0x1999
	ErrBadMechanism errors.Error = 0x1999<<8 + iota
	ErrTooShortLifetime
	ErrLocalAddressNotFound
	ErrAddMapping
	ErrNoRouterDiscovered
)

var errMsgs = errors.Messages{
	ErrBadMechanism:         "bad mechanism",
	ErrTooShortLifetime:     "too short lifetime",
	ErrLocalAddressNotFound: "failed to find local address",
	ErrAddMapping:           "failed to add port mapping",
	ErrNoRouterDiscovered:   "no router discovered",
}

func init() { errors.InjectMessages(errMsgs) }
