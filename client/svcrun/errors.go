package svcrun

import "github.com/privatix/dappctrl/util/errors"

// ServiceRunner errors.
const (
	// CRC16("github.com/privatix/dappctrl/client/svcrun") = 0x0214
	ErrAlreadyStarted errors.Error = 0x0214<<8 + iota
	ErrUnknownService
	ErrNotRunning
)

func init() {
	errors.InjectMessages(map[errors.Error]string{
		ErrAlreadyStarted: "service already running",
		ErrUnknownService: "unknown service type",
		ErrNotRunning:     "not running",
	})
}
