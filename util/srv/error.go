package srv

import (
	"fmt"
	"net/http"
)

// Error is a server error.
type Error struct {
	status  int
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf(
		"server responed with error: %s (%d)", e.Message, e.Code)
}

// Common server error codes.
const (
	ErrCodeMethodNotAllowed     = 1
	ErrCodeInternalServerError  = iota
	ErrCodeFailedToParseRequest = iota
	ErrCodeAccessDenied         = iota
	ErrCodeMax                  = 100
)

// Common server errors.
var (
	ErrMethodNotAllowed = &Error{
		status:  http.StatusMethodNotAllowed,
		Code:    ErrCodeMethodNotAllowed,
		Message: "HTTP method not allowed",
	}
	ErrInternalServerError = &Error{
		status:  http.StatusInternalServerError,
		Code:    ErrCodeInternalServerError,
		Message: "internal server error",
	}
	ErrFailedToParseRequest = &Error{
		status:  http.StatusBadRequest,
		Code:    ErrCodeFailedToParseRequest,
		Message: "failed to parse request",
	}
	ErrAccessDenied = &Error{
		status:  http.StatusForbidden,
		Code:    ErrCodeAccessDenied,
		Message: "access denied",
	}
)
