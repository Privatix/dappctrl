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
	ErrChannelNotFound
	ErrDefailtGasPriceNotFound
	ErrOfferingNotFound
	ErrBadObjectType
)

var errMsgs = errors.Messages{
	ErrAccessDenied:            "access denied",
	ErrInternal:                "internal server error",
	ErrAccountNotFound:         "account not found",
	ErrChannelNotFound:         "channel not found",
	ErrDefailtGasPriceNotFound: "default gas price setting not found",
	ErrOfferingNotFound:        "offering not found",
	ErrBadObjectType:           "bad object type",
}

func init() { errors.InjectMessages(errMsgs) }
