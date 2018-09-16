package ui

import (
	"github.com/privatix/dappctrl/util/errors"
)

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/ui") = 0x2F5D
	ErrAccessDenied errors.Error = 0x2F5D<<8 + iota
	ErrInternal
	ErrObjectNotFound
	ErrAccountNotFound
	ErrChannelNotFound
	ErrDefailtGasPriceNotFound
	ErrEmptyPassword
	ErrOfferingNotFound
	ErrSettingNotFound
	ErrBadObjectType
	ErrPasswordExists
	ErrProductNotFound
	ErrInvalidTemplateType
	ErrMalformedTemplate
)

var errMsgs = errors.Messages{
	ErrAccessDenied:            "access denied",
	ErrInternal:                "internal server error",
	ErrObjectNotFound:          "object not found",
	ErrAccountNotFound:         "account not found",
	ErrChannelNotFound:         "channel not found",
	ErrDefailtGasPriceNotFound: "default gas price setting not found",
	ErrEmptyPassword:           "invalid password",
	ErrOfferingNotFound:        "offering not found",
	ErrSettingNotFound:         "setting not found",
	ErrBadObjectType:           "bad object type",
	ErrPasswordExists:          "password exists",
	ErrProductNotFound:         "product not found",
	ErrInvalidTemplateType:     "invalid template type",
	ErrMalformedTemplate:       "malformed template",
}

func init() { errors.InjectMessages(errMsgs) }
