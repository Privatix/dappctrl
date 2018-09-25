// +build !notest

package country

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

// ServerMock it is mock for processing a request for country detection.
type ServerMock struct {
	Server *httptest.Server
}

// NewServerMock creates new ServerMock.
func NewServerMock(field, result string) *ServerMock {
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, fmt.Sprintf(`{"%s": "%s"}`,
				field, result))
		}))
	return &ServerMock{ts}
}

// Close closed ServerMock.
func (s ServerMock) Close() {
	s.Server.Close()
}
