package worker

import "errors"

// Errors returned by workers.
var (
	ErrInvalidJob = errors.New("invalid job definition")
)
