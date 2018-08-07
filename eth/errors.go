package eth

import "github.com/privatix/dappctrl/util/errors"

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/eth") = 0x82E7
	ErrURLScheme errors.Error = 0x82E7<<8 + iota
	ErrCreateClient
)

var errMsgs = errors.Messages{
	ErrURLScheme:    "no known transport for URL scheme",
	ErrCreateClient: "failed to create rpc client ",
}

func init() { errors.InjectMessages(errMsgs) }
