// +build !notest

package job

import (
	"github.com/privatix/dappctrl/data"
)

// Mock methods.
const (
	MockAdd = iota
	MockProcess
	MockClose
	MockSubscribe
	MockUnsubscribe
)

// QueueMock is a queue method handler.
type QueueMock func(method int, job *data.Job,
	relatedIDs []string, subID string, subFunc SubFunc) error

// NewDummyQueueMock returns a queue mock which does nothing.
func NewDummyQueueMock() QueueMock {
	return func(method int, job *data.Job,
		relatedIDs []string, subID string, subFunc SubFunc) error {
		return nil
	}
}

// Add is a mock implementation for the Add queue method.
func (q QueueMock) Add(j *data.Job) error { return q(MockAdd, j, nil, "", nil) }

// Process is a mock implementation for the Process queue method.
func (q QueueMock) Process() error { return q(MockProcess, nil, nil, "", nil) }

// Close is a mock implementation for the Close queue method.
func (q QueueMock) Close() { q(MockClose, nil, nil, "", nil) }

// Subscribe is a mock implementation for the Subscribe queue method.
func (q QueueMock) Subscribe(
	relatedIDs []string, subID string, subFunc SubFunc) error {
	return q(MockSubscribe, nil, relatedIDs, subID, subFunc)
}

// Unsubscribe is a mock implementation for the Unsubscribe queue method.
func (q QueueMock) Unsubscribe(relatedIDs []string, subID string) error {
	return q(MockUnsubscribe, nil, relatedIDs, subID, nil)
}
