package statik

import (
	"github.com/privatix/dappctrl/util/errors"
)

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/statik") = 0xBB3D
	ErrOpenFS errors.Error = 0xBB3D<<8 + iota
	ErrOpenFile
	ErrReadFile
)

var errMsgs = errors.Messages{
	ErrOpenFS:   "failed to open statik filesystem",
	ErrOpenFile: "failed to open statik file",
	ErrReadFile: "failed to read statik file",
}

func init() { errors.InjectMessages(errMsgs) }
