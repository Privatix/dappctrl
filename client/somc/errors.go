package somc

import (
	"github.com/privatix/dappctrl/util/errors"
)

// Errors returned by workers.
const (
	// CRC16("github.com/privatix/dappctrl/client/somc") = 0x42AE
	ErrUnknownSOMCType errors.Error = 0x42AE<<8 + iota
)

var errMsgs = errors.Messages{
	ErrUnknownSOMCType: "unknown somc type",
}

func init() {
	errors.InjectMessages(errMsgs)
}
