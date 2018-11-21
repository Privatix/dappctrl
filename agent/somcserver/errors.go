package somcserver

import "github.com/privatix/dappctrl/util/errors"

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/agent/somcserver") = 0x9B09
	ErrInternal errors.Error = 0x9B09<<8 + iota
	ErrChannelNotFound
	ErrEndpointNotFound
	ErrOfferingNotFound
)

var errMsgs = errors.Messages{
	ErrInternal:         "internal error occurred",
	ErrChannelNotFound:  "channel not found",
	ErrEndpointNotFound: "endpoint not found",
	ErrOfferingNotFound: "offering not found",
}

func init() { errors.InjectMessages(errMsgs) }
