package country

import "github.com/privatix/dappctrl/util/errors"

// Errors returned by workers.
const (
	ErrInternal errors.Error = 0xFFDA<<8 + iota
	ErrMissingRequiredField
	ErrBadCountryValueType
)

var errMsgs = errors.Messages{
	ErrInternal:             "internal server error",
	ErrMissingRequiredField: "missing required field",
	ErrBadCountryValueType:  "country value is not a string",
}

func init() {
	errors.InjectMessages(errMsgs)
}
