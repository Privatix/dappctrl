package sesssrv

import "github.com/privatix/dappctrl/util/errors"

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/sesssrv") = 0x07C8
	ErrEncodeArgs errors.Error = 0x07C8<<8 + iota
	ErrDecodeResponse
)

var errMsgs = errors.Messages{
	ErrEncodeArgs:     "failed to encode arguments to JSON",
	ErrDecodeResponse: "failed to decode response from JSON",
}

func init() { errors.InjectMessages(errMsgs) }
