package worker

import "errors"

// Errors returned by workers.
var (
	ErrInvalidJob       = errors.New("unexpected job type or job related type")
	ErrNotEnoughBalance = errors.New("not enough PSC balance for offering")
	ErrChReceiptBalance = errors.New("receipt balance is greater than a deposit")
	ErrInvalidChStatus  = errors.New("can not be applied to a channel with" +
		" the current channel status")
	ErrInvalidServiceStatus = errors.New("can not be applied to a channel with" +
		" the current service status")
	ErrNoSupply = errors.New("no supply")
	ErrBadServiceStatus = errors.New("bad service status")
)
