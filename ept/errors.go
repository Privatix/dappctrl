package ept

import "github.com/pkg/errors"

// Endpoint Message Template errors
var (
	ErrInput             = errors.New("one or more input parameters is wrong")
	ErrTimeOut           = errors.New("timeout")
	ErrInvalidFormat     = errors.New("invalid endpoint message format")
	ErrProdOfferAccessID = errors.New("OfferAccessID from product is null")
)

func errWrapper(err error, msg string) error {
	return errors.Wrap(err, msg)
}
