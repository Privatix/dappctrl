package ui

import (
	"github.com/privatix/dappctrl/util/errors"
)

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/ui") = 0x2F5D
	ErrAccessDenied errors.Error = 0x2F5D<<8 + iota
	ErrAccountNotFound
	ErrBadObjectType
	ErrChannelNotFound
	ErrDefailtGasPriceNotFound
	ErrEmptyPassword
	ErrInternal
	ErrOfferingNotFound
	ErrPasswordExists
	ErrProductNotFound
)

var errMsgs = errors.Messages{
	ErrAccessDenied:            "access denied",
	ErrAccountNotFound:         "account not found",
	ErrBadObjectType:           "bad object type",
	ErrChannelNotFound:         "channel not found",
	ErrDefailtGasPriceNotFound: "default gas price setting not found",
	ErrEmptyPassword:           "invalid password",
	ErrInternal:                "internal server error",
	ErrOfferingNotFound:        "offering not found",
	ErrPasswordExists:          "password exists",
	ErrProductNotFound:         "product not found",
}

func init() { errors.InjectMessages(errMsgs) }
