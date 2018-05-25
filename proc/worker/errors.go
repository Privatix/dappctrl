package worker

import "errors"

// Errors returned by workers.
var (
	ErrInvalidJob = errors.New("unexpected job type or job related type")
)
