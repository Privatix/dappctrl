package log

import (
	"github.com/privatix/dappctrl/util/errors"
)

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/util/log") = 0x6928
	ErrBadLevel errors.Error = 0x6928<<8 + iota
	ErrBadStackLevel
)

var errMsgs = errors.Messages{
	ErrBadLevel:      "bad log level",
	ErrBadStackLevel: "bad log level for stack trace",
}

func init() { errors.InjectMessages(errMsgs) }
