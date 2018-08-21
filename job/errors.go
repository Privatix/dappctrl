package job

import "github.com/privatix/dappctrl/util/errors"

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/job") = 0x765D
	ErrAlreadyProcessing errors.Error = 0x765D<<8 + iota
	ErrDuplicatedJob
	ErrHandlerNotFound
	ErrQueueClosed
	ErrSubscriptionExists
	ErrSubscriptionNotFound
	ErrInternal
)

var errMsgs = errors.Messages{
	ErrAlreadyProcessing:    "already processing",
	ErrDuplicatedJob:        "duplicated job",
	ErrHandlerNotFound:      "job handler not found",
	ErrQueueClosed:          "queue closed",
	ErrSubscriptionExists:   "subscription already exists",
	ErrSubscriptionNotFound: "subscription not found",
	ErrInternal:             "internal server error",
}

func init() { errors.InjectMessages(errMsgs) }
