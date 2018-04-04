// +build !nosomctest

package somc

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/privatix/dappctrl/data"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/util"
)

var conf struct {
	Log      *util.LogConfig
	SOMC     *Config
	SOMCTest *TestConfig
}

var logger *util.Logger

func newServer(t *testing.T) *FakeSOMC {
	return NewFakeSOMC(t, conf.SOMC.URL, conf.SOMCTest.ServerStartupDelay)
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
	defer srv.Close()
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
	defer srv.Close()
	conn := newConn(t)
	defer conn.Close()

	ch := make(chan error)
	go func() {
		ch <- conn.PublishOffering([]byte("{}"))
	}()

	req := srv.Read(t, publishOfferingMethod)
	repl := JSONRPCMessage{ID: req.ID, Result: []byte("true")}
	srv.Write(t, &repl)

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
	defer srv.Close()
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

	req := srv.Read(t, findOfferingsMethod)
	repl := JSONRPCMessage{ID: req.ID, Result: data}
	srv.Write(t, &repl)

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

	req = srv.Read(t, findOfferingsMethod)
	repl = JSONRPCMessage{ID: req.ID, Result: data}
	srv.Write(t, &repl)

	if ret := <-ch; ret.err == nil {
		t.Fatalf("hash mismatch error expected, but not occurred")
	}
}

func TestPublishEndpoint(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()
	conn := newConn(t)
	defer conn.Close()

	ch := make(chan error)
	go func() {
		ch <- conn.PublishEndpoint("a", []byte("{}"))
	}()

	req := srv.Read(t, publishEndpointMethod)
	repl := JSONRPCMessage{ID: req.ID, Result: []byte("true")}
	srv.Write(t, &repl)

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
	defer srv.Close()
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
		req := srv.Read(t, waitForEndpointMethod)
		repl := JSONRPCMessage{ID: req.ID, Result: []byte("true")}
		srv.Write(t, &repl)
	}

	time.Sleep(time.Millisecond)

	params := endpointParams{Channel: "a", Endpoint: []byte("{}")}
	data, _ := json.Marshal(&params)

	repl := JSONRPCMessage{
		ID:     100,
		Method: publishEndpointMethod,
		Params: data,
	}
	srv.Write(t, &repl)

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
	conf.SOMCTest = NewTestConfig()
	util.ReadTestConfig(&conf)

	logger = util.NewTestLogger(conf.Log)

	os.Exit(m.Run())
}
