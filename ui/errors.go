package ui

import (
	"github.com/privatix/dappctrl/util/errors"
)

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/ui") = 0x2F5D
	ErrAccessDenied errors.Error = 0x2F5D<<8 + iota
	ErrInternal
	ErrAccountNotFound
	ErrOfferingNotFound
)

var errMsgs = errors.Messages{
	ErrAccessDenied:     "access denied",
	ErrInternal:         "internal server error",
	ErrAccountNotFound:  "account not found",
	ErrOfferingNotFound: "offering not found",
}

func init() { errors.InjectMessages(errMsgs) }
