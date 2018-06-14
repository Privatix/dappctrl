package worker

import "errors"

// Errors returned by workers.
var (
	ErrInvalidJob       = errors.New("unexpected job type or job related type")
	ErrNotEnoughBalance = errors.New("not enough PSC balance for offering")
	ErrNoSupply         = errors.New("no supply")
)
