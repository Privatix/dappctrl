// +build !notest

package svcrun

// Mock methods.
const (
	MockStart     = iota
	MockIsRunning = iota
	MockStop      = iota
	MockStopAll   = iota
)

// Mock is a ServiceRunner method handler.
type Mock func(method int, channel string) (bool, error)

// NewIdleMock returns a mock which does nothing.
func NewIdleMock() Mock {
	return func(method int, channel string) (bool, error) {
		return false, nil
	}
}

// Start is a mock implementation for the Start service runner method.
func (m Mock) Start(channel string) error {
	_, err := m(MockStart, channel)
	return err
}

// IsRunning is a mock implementation for the IsRunning service runner method.
func (m Mock) IsRunning(channel string) (bool, error) {
	return m(MockIsRunning, channel)
}

// Stop is a mock implementation for the Stop service runner method.
func (m Mock) Stop(channel string) error {
	_, err := m(MockStop, channel)
	return err
}

// StopAll is a mock implementation for the StopAll service runner method.
func (m Mock) StopAll() error {
	_, err := m(MockStopAll, "")
	return err
}
