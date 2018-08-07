package ept

import (
	"github.com/privatix/dappctrl/util/errors"
)

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/messages/ept") = 0xDC36
	ErrTimeOut errors.Error = 0xDC36<<8 + iota
	ErrInvalidFormat
	ErrProdOfferAccessID
	ErrProdEndAddress
)

var errMsgs = errors.Messages{
	ErrTimeOut:           "timeout",
	ErrInvalidFormat:     "invalid endpoint message format",
	ErrProdOfferAccessID: "OfferAccessID from product is null",
	ErrProdEndAddress:    "ServiceEndpointAddress from product is null",
}

func init() { errors.InjectMessages(errMsgs) }
