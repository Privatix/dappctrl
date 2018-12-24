package somc

import "github.com/privatix/dappctrl/util/errors"

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/agent/somc") = 0xC4C6
	ErrNoTorHostname errors.Error = 0xC4C6<<8 + iota
	ErrNoDirectAddr
	ErrNoActiveTransport
	ErrUnknownSOMCType
)

var errMsgs = errors.Messages{
	ErrNoTorHostname:     "incomplete config: tor hostname not set",
	ErrNoDirectAddr:      "incomplete config: direct address not set",
	ErrNoActiveTransport: "none of SOMC's are active",
	ErrUnknownSOMCType:   "unknown SOMC type",
}

func init() { errors.InjectMessages(errMsgs) }
