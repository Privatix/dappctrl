package errors

import (
	"fmt"
)

// Error is an error code supporting the standard error interface.
type Error int

// Error returns an error message of a given error.
func (e Error) Error() string {
	if msg, ok := Message(e); ok {
		return fmt.Sprintf("%s (%d)", msg, e)
	}
	return "unknown error"
}

// Code returns an error code of a given error.
func (e Error) Code() int {
	return int(e)
}

// Messages is a mapping between error codes and error messages.
type Messages map[Error]string

var msgs = Messages{}

// InjectMessages injects errors messages into a global message map. Any
// package with own errors defined should call this function during its
// initialisation (i.e. in init() function).
func InjectMessages(m Messages) {
	for k, v := range m {
		if _, ok := msgs[k]; ok {
			panic(fmt.Sprintf("duplicated error: %d", k))
		}
		msgs[k] = v
	}
}

// Message returns an error message from a given error code.
func Message(e Error) (string, bool) {
	msg, ok := msgs[e]
	return msg, ok
}
