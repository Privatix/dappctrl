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
	ErrMinConfirmationsNotFound
	ErrOfferingNotFound
	ErrPasswordExists
	ErrProductNotFound
	ErrInvalidTemplateType
	ErrMalformedTemplate
	ErrDepositTooSmall
	ErrBadUnitPriceRange
	ErrBadUnitType
	ErrBillingType
	ErrBadOfferingStatusAction
	ErrBadObjectType
	ErrFailedToDecodePrivateKey
	ErrFailedToDecryptPKey
	ErrPrivateKeyNotFound
	ErrTokenAmountTooSmall
	ErrBadDestination
)

var errMsgs = errors.Messages{
	ErrAccessDenied:             "access denied",
	ErrInternal:                 "internal server error",
	ErrObjectNotFound:           "object not found",
	ErrAccountNotFound:          "account not found",
	ErrChannelNotFound:          "channel not found",
	ErrDefailtGasPriceNotFound:  "default gas price setting not found",
	ErrEmptyPassword:            "invalid password",
	ErrMinConfirmationsNotFound: "min confirmations setting not found",
	ErrOfferingNotFound:         "offering not found",
	ErrBadObjectType:            "bad object type",
	ErrDepositTooSmall:          "deposit is too small",
	ErrBadUnitPriceRange:        "bad unit price range",
	ErrBadUnitType:              "bad unit type",
	ErrBillingType:              "bad billing type",
	ErrBadOfferingStatusAction:  "bad offering status action",
	ErrPasswordExists:           "password exists",
	ErrProductNotFound:          "product not found",
	ErrInvalidTemplateType:      "invalid template type",
	ErrMalformedTemplate:        "malformed template",
	ErrFailedToDecodePrivateKey: "failed to decode private key",
	ErrFailedToDecryptPKey:      "failed to decrypt private key from json blob",
	ErrPrivateKeyNotFound:       "private key not found",
	ErrTokenAmountTooSmall:      "the amount of tokens is too small",
	ErrBadDestination:           "bad destination",
}

func init() { errors.InjectMessages(errMsgs) }
