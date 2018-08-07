package rpcsrv

import (
	"github.com/ethereum/go-ethereum/rpc"
)

// Error is an error to be passed within RPC notification. Regular errors
// cannot be properly marshaled into JSON, so they should be converted into
// Error pointers before being sent.
type Error struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// ToError converts error into Error pointer.
func ToError(err error) *Error {
	if err == nil {
		return nil
	}

	if err, ok := err.(rpc.Error); ok {
		return &Error{Code: err.ErrorCode(), Message: err.Error()}
	}

	return &Error{Message: err.Error()}
}

// Error returns an error message of a given error.
func (e *Error) Error() string { return e.Message }

// ErrorCode returns an error code of a given error.
func (e *Error) ErrorCode() int { return e.Code }
