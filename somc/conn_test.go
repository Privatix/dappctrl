// +build !nosomctest

package somc

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/privatix/dappctrl/data"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/websocket"

	"github.com/privatix/dappctrl/util"
)

type somcTestConfig struct {
	ServerStartupDelay uint // In milliseconds.
}

func newSOMCTestConfig() *somcTestConfig {
	return &somcTestConfig{
		ServerStartupDelay: 10,
	}
}

var conf struct {
	Log      *util.LogConfig
	SOMC     *Config
	SOMCTest *somcTestConfig
}

var logger *util.Logger

type server struct {
	sync.Mutex // To make sure conn cannot be implicitly rewritten.

	srv  *http.Server
	conn *websocket.Conn
}

func newServer(t *testing.T) *server {
	mux := http.NewServeMux()

	sp := strings.Split(conf.SOMC.URL, "/")
	if len(sp) < 3 {
		t.Fatalf("bad SOMC URL: %s", conf.SOMC.URL)
	}

	srv := &server{srv: &http.Server{Addr: sp[2], Handler: mux}}

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

	time.Sleep(time.Duration(conf.SOMCTest.ServerStartupDelay) *
		time.Millisecond)

	return srv
}

func (s *server) close() {
	if s.conn != nil {
		s.conn.Close()
	}
	s.srv.Close()
}

func (s *server) read(t *testing.T, method string) *jsonRPCMessage {
	var msg jsonRPCMessage
	if err := s.conn.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read message: %s", err)
	}

	if msg.Version != jsonRPCVersion || msg.Method != method {
		t.Fatalf("bad message format")
	}

	return &msg
}

func (s *server) write(t *testing.T, msg *jsonRPCMessage) {
	if err := s.conn.WriteJSON(msg); err != nil {
		t.Fatalf("failed to write message: %s", err)
	}
}

func newConn(t *testing.T) *Conn {
	conn, err := NewConn(conf.SOMC, logger)
	if err != nil {
		t.Fatalf("failed to create connection: %s", err)
	}
	return conn
}

func TestReconnect(t *testing.T) {
	srv := newServer(t)
	defer srv.close()
	conn := newConn(t)
	defer conn.Close()

	ch := make(chan error)
	go func() {
		ch <- conn.PublishOffering([]byte("{}"))
	}()

	srv.conn.Close()

	if err := <-ch; err == nil {
		t.Fatalf("disconnect error expected, but not occurred")
	}

	srv.conn = nil

	for i := 0; i < int(time.Second/time.Millisecond); i++ {
		if srv.conn != nil {
			break
		}
		time.Sleep(time.Millisecond)
	}

	if srv.conn == nil {
		t.Fatalf("failed to reconnect")
	}
}

func TestPublishOffering(t *testing.T) {
	srv := newServer(t)
	defer srv.close()
	conn := newConn(t)
	defer conn.Close()

	ch := make(chan error)
	go func() {
		ch <- conn.PublishOffering([]byte("{}"))
	}()

	req := srv.read(t, publishOfferingMethod)
	repl := jsonRPCMessage{ID: req.ID, Result: []byte("true")}
	srv.write(t, &repl)

	if err := <-ch; err != nil {
		t.Fatalf("failed to publish offering: %s", err)
	}
}

type findOfferingsReturn struct {
	data []OfferingData
	err  error
}

func TestFindOffering(t *testing.T) {
	srv := newServer(t)
	defer srv.close()
	conn := newConn(t)
	defer conn.Close()

	off := []byte("{}")
	hash := crypto.Keccak256Hash(off)
	hstr := data.FromBytes(hash.Bytes())

	ch := make(chan findOfferingsReturn)
	go func() {
		data, err := conn.FindOfferings([]string{hstr})
		ch <- findOfferingsReturn{data, err}
	}()

	ostr := data.FromBytes(off)
	res := findOfferingsResult{{Hash: hstr, Data: ostr}}
	data, _ := json.Marshal(res)

	req := srv.read(t, findOfferingsMethod)
	repl := jsonRPCMessage{ID: req.ID, Result: data}
	srv.write(t, &repl)

	ret := <-ch

	if ret.err != nil {
		t.Fatalf("failed to find offerings: %s", ret.err)
	}

	if len(ret.data) != 1 || bytes.Compare(ret.data[0].Offering, off) != 0 {
		t.Fatalf("offering data mismatch")
	}

	go func() {
		data, err := conn.FindOfferings([]string{hstr})
		ch <- findOfferingsReturn{data, err}
	}()

	// Try the same but with a wrong hash.
	res[0].Hash = "x" + res[0].Hash[1:]
	data, _ = json.Marshal(res)

	req = srv.read(t, findOfferingsMethod)
	repl = jsonRPCMessage{ID: req.ID, Result: data}
	srv.write(t, &repl)

	if ret := <-ch; ret.err == nil {
		t.Fatalf("hash mismatch error expected, but not occurred")
	}
}

func TestPublishEndpoint(t *testing.T) {
	srv := newServer(t)
	defer srv.close()
	conn := newConn(t)
	defer conn.Close()

	ch := make(chan error)
	go func() {
		ch <- conn.PublishEndpoint("a", []byte("{}"))
	}()

	req := srv.read(t, publishEndpointMethod)
	repl := jsonRPCMessage{ID: req.ID, Result: []byte("true")}
	srv.write(t, &repl)

	if err := <-ch; err != nil {
		t.Fatalf("failed to publish endpoint: %s", err)
	}
}

type waitForEndpointReturn struct {
	data []byte
	err  error
}

func TestWaitForEndpoint(t *testing.T) {
	srv := newServer(t)
	defer srv.close()
	conn := newConn(t)
	defer conn.Close()

	ch := make(chan waitForEndpointReturn)
	for i := 0; i < 2; i++ {
		go func() {
			data, err := conn.WaitForEndpoint("a")
			ch <- waitForEndpointReturn{data, err}
		}()
	}

	for i := 0; i < 2; i++ {
		req := srv.read(t, waitForEndpointMethod)
		repl := jsonRPCMessage{ID: req.ID, Result: []byte("true")}
		srv.write(t, &repl)
	}

	time.Sleep(time.Millisecond)

	params := endpointParams{Channel: "a", Endpoint: []byte("{}")}
	data, _ := json.Marshal(&params)

	repl := jsonRPCMessage{
		ID:     100,
		Method: publishEndpointMethod,
		Params: data,
	}
	srv.write(t, &repl)

	for i := 0; i < 2; i++ {
		ret := <-ch

		if ret.err != nil {
			t.Fatalf("failed to get endpoint data: %s", ret.err)
		}

		if bytes.Compare(ret.data, []byte("{}")) != 0 {
			t.Fatalf("endpoint data mismatch")
		}
	}
}

func TestMain(m *testing.M) {
	conf.Log = util.NewLogConfig()
	conf.SOMC = NewConfig()
	conf.SOMCTest = newSOMCTestConfig()
	util.ReadTestConfig(&conf)

	logger = util.NewTestLogger(conf.Log)

	os.Exit(m.Run())
}
