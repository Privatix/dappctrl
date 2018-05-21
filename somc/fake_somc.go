// +build !notest

package somc

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestEndpointParams exported for tests.
type TestEndpointParams endpointParams

// TestOfferingParams used for tests.
type TestOfferingParams publishOfferingParams

// TestConfig the config related to somc tests.
type TestConfig struct {
	ServerStartupDelay uint // In milliseconds.
}

// NewTestConfig returns default test config.
func NewTestConfig() *TestConfig {
	return &TestConfig{
		ServerStartupDelay: 10,
	}
}

// FakeSOMC is a fake somc server.
type FakeSOMC struct {
	sync.Mutex // To make sure conn cannot be implicitly rewritten.

	srv  *http.Server
	conn *websocket.Conn
}

// NewFakeSOMC returns new fake somc server.
func NewFakeSOMC(t *testing.T, somcURL string, startupDelay uint) *FakeSOMC {
	mux := http.NewServeMux()

	sp := strings.Split(somcURL, "/")
	if len(sp) < 3 {
		t.Fatalf("bad SOMC URL: %s", somcURL)
	}

	srv := &FakeSOMC{srv: &http.Server{Addr: sp[2], Handler: mux}}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		srv.Lock()
		defer srv.Unlock()

		if srv.conn != nil {
			return
		}

		up := websocket.Upgrader{}

		conn, err := up.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("failed to upgrade: %s", err)
			return
		}

		srv.conn = conn
	})

	go func() {
		err := srv.srv.ListenAndServe()
		if err != http.ErrServerClosed {
			t.Fatalf("failed to listen and serve: %s", err)
		}
	}()

	time.Sleep(time.Duration(startupDelay) * time.Millisecond)

	return srv
}

// Close closes connection and stops server.
func (s *FakeSOMC) Close() {
	if s.conn != nil {
		s.conn.Close()
	}
	s.srv.Close()
}

// Read reads from connection.
func (s *FakeSOMC) Read(t *testing.T, method string) *JSONRPCMessage {
	var msg JSONRPCMessage
	if err := s.conn.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read message: %s", err)
	}

	if msg.Version != jsonRPCVersion || msg.Method != method {
		t.Fatalf("bad message format")
	}

	return &msg
}

// Write writes reply to connection.
func (s *FakeSOMC) Write(t *testing.T, msg *JSONRPCMessage) {
	if err := s.conn.WriteJSON(msg); err != nil {
		t.Fatalf("failed to write message: %s", err)
	}
}

// ReadPublishEndpoint recieves and returns published endpoint.
func (s *FakeSOMC) ReadPublishEndpoint(t *testing.T) TestEndpointParams {
	req := s.Read(t, publishEndpointMethod)
	params := endpointParams{}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		t.Fatal("FakeSOMC: failed to unmurshal params: ", err)
	}
	repl := JSONRPCMessage{ID: req.ID, Result: []byte("true")}
	s.Write(t, &repl)
	return TestEndpointParams(params)
}

// ReadPublishOfferings recieves and returns published endpoint.
func (s *FakeSOMC) ReadPublishOfferings(t *testing.T) TestOfferingParams {
	req := s.Read(t, publishOfferingMethod)
	params := publishOfferingParams{}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		t.Fatal("FakeSOMC: failed to unmurshal params: ", err)
	}
	repl := JSONRPCMessage{ID: req.ID, Result: []byte("true")}
	s.Write(t, &repl)
	return TestOfferingParams(params)
}
