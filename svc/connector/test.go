package connector

import (
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/sesssrv"
)

// Mock is a mock implementation for Connector interface.
type Mock struct {
	Error    error
	Endpoint *data.Endpoint
}

// NewMock returns mock for Connector interface.
func NewMock() *Mock {
	return &Mock{}
}

// AuthSession is a mock implementation for the AuthSession connector method.
func (m *Mock) AuthSession(args *sesssrv.AuthArgs) error {
	return m.Error
}

// StartSession is a mock implementation for the StartSession connector method.
func (m *Mock) StartSession(
	args *sesssrv.StartArgs) (*sesssrv.StartResult, error) {
	return nil, m.Error
}

// StopSession is a mock implementation for the StopSession connector method.
func (m *Mock) StopSession(args *sesssrv.StopArgs) error {
	return m.Error
}

// UpdateSessionUsage is a mock implementation
// for the UpdateSessionUsage connector method.
func (m *Mock) UpdateSessionUsage(args *sesssrv.UpdateArgs) error {
	return m.Error
}

// SetupProductConfiguration is a mock implementation
// for the SetupProductConfiguration connector method.
func (m *Mock) SetupProductConfiguration(args *sesssrv.ProductArgs) error {
	return m.Error
}

// GetEndpointMessage is a mock implementation
// for the GetEndpointMessage connector method.
func (m *Mock) GetEndpointMessage(
	args *sesssrv.EndpointMsgArgs) (*data.Endpoint, error) {
	return m.Endpoint, m.Error
}
