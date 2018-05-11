package ept

import "github.com/pkg/errors"

// Endpoint Message Template errors
var (
	ErrTimeOut                = errors.New("timeout")
	ErrInvalidEndpointMessage = errors.New("invalid endpoint message")
)
