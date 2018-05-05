package sesssrv

import (
	"github.com/privatix/dappctrl/util/srv"
)

// Common error codes for session server.
const (
	ErrCodeChannelNotFound            = srv.ErrCodeMax + 1
	ErrCodeBadAuthPassword            = iota
	ErrCodeNonActiveChannel           = iota
	ErrCodeSessionNotFound            = iota
	ErrCodeProductConfAlreadyUploaded = iota
	ErrCodeProductConfNotValid        = iota
)

// Common session server errors.
var (
	ErrChannelNotFound = &srv.Error{
		Code:    ErrCodeChannelNotFound,
		Message: "channel not found",
	}
	ErrBadAuthPassword = &srv.Error{
		Code:    ErrCodeBadAuthPassword,
		Message: "bad authentication password",
	}
	ErrNonActiveChannel = &srv.Error{
		Code:    ErrCodeNonActiveChannel,
		Message: "non-active channel",
	}
	ErrSessionNotFound = &srv.Error{
		Code:    ErrCodeSessionNotFound,
		Message: "session not found",
	}
	ErrProductConfAlreadyUploaded = &srv.Error{
		Code:    ErrCodeProductConfAlreadyUploaded,
		Message: "product configuration already uploaded",
	}
	ErrProductConfNotValid = &srv.Error{
		Code:    ErrCodeProductConfNotValid,
		Message: "product configuration not valid",
	}
)
