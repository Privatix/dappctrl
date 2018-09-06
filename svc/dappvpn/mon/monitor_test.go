// +build !nosvcdappvpnmontest

package mon

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

type testConfig struct {
	ServerStartupDelay uint // In milliseconds.
}

func newTestConfig() *testConfig {
	return &testConfig{
		ServerStartupDelay: 10,
	}
}

var conf struct {
	FileLog        *log.FileConfig
	VPNMonitor     *Config
	VPNMonitorTest *testConfig
}

var logger log.Logger

func connect(t *testing.T, handleSession HandleSessionFunc,
	channel string) (net.Conn, <-chan error) {
	lst, err := net.Listen("tcp", conf.VPNMonitor.Addr)
	if err != nil {
		t.Fatalf("failed to listen: %s", err)
	}
	defer lst.Close()

	time.Sleep(time.Duration(conf.VPNMonitorTest.ServerStartupDelay) *
		time.Millisecond)

	ch := make(chan error)
	go func() {
		mon := NewMonitor(
			conf.VPNMonitor, logger, handleSession, channel)
		ch <- mon.MonitorTraffic()
		mon.Close()
	}()

	var conn net.Conn
	if conn, err = lst.Accept(); err != nil {
		t.Fatalf("failed to accept: %s", err)
	}

	return conn, ch
}

func expectExit(t *testing.T, ch <-chan error, expected error) {
	err := <-ch

	_, neterr := err.(net.Error)
	disconn := neterr || err == io.EOF

	if (disconn && expected != nil) || (!disconn && err != expected) {
		t.Fatalf("unexpected monitor error: %s", err)
	}
}

func exit(t *testing.T, conn net.Conn, ch <-chan error) {
	conn.Close()
	expectExit(t, ch, nil)
}

func send(t *testing.T, conn net.Conn, str string) {
	if _, err := conn.Write([]byte(str + "\n")); err != nil {
		t.Fatalf("failed to send to monitor: %s", err)
	}
}

func receive(t *testing.T, reader *bufio.Reader) string {
	str, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("failed to receive from monitor: %s", err)
	}
	return strings.TrimRight(str, "\r\n")
}

func assertNothingToReceive(t *testing.T, conn net.Conn, reader *bufio.Reader) {
	conn.SetReadDeadline(time.Now().Add(time.Millisecond))

	str, err := reader.ReadString('\n')
	if err == nil {
		t.Fatalf("unexpected message received: %s", str)
	}

	if neterr, ok := err.(net.Error); !ok || !neterr.Timeout() {
		t.Fatalf("non-timeout error: %s", err)
	}
}

func handleSessionEvent(ch string, event int, up, down uint64) bool {
	return true
}

func TestOldOpenVPN(t *testing.T) {
	conn, ch := connect(t, handleSessionEvent, "")
	defer conn.Close()

	send(t, conn, prefixClientListHeader)
	send(t, conn, prefixClientList+",,,,,,,,")

	expectExit(t, ch, ErrServerOutdated)
}

func checkByteCount(t *testing.T, reader *bufio.Reader) {
	cmd := fmt.Sprintf("bytecount %d", conf.VPNMonitor.ByteCountPeriod)
	if str := receive(t, reader); str != cmd {
		t.Fatalf("unexpected bytecount command: %s", str)
	}
}

func TestInitFlow(t *testing.T) {
	conn, ch := connect(t, handleSessionEvent, "")
	defer conn.Close()

	reader := bufio.NewReader(conn)

	checkByteCount(t, reader)

	if str := receive(t, reader); str != "status 2" {
		t.Fatalf("unexpected status command: %s", str)
	}

	exit(t, conn, ch)
}

const (
	cid         = 0
	up, down    = 1024, 2048
	commonName  = "Common-Name"
	testChannel = "Test-Channel"
)

func TestClientInitFlow(t *testing.T) {
	conn, ch := connect(t, handleSessionEvent, testChannel)
	defer conn.Close()

	reader := bufio.NewReader(conn)

	checkByteCount(t, reader)

	if str := receive(t, reader); str != "state on" {
		t.Fatalf("unexpected state command: %s", str)
	}

	if str := receive(t, reader); str != "hold release" {
		t.Fatalf("unexpected hold command: %s", str)
	}

	exit(t, conn, ch)
}

func sendByteCount(t *testing.T, conn net.Conn) {
	send(t, conn, prefixClientListHeader)
	send(t, conn, fmt.Sprintf("%s%s,,,,,,,,%s,%d",
		prefixClientList, commonName, testChannel, cid))
	send(t, conn, fmt.Sprintf("%s%d,%d,%d", prefixByteCount, cid, down, up))
}

func sendByteCountClient(t *testing.T, conn net.Conn) {
	msg := fmt.Sprintf("%s%d,%d", prefixByteCountClient, down, up)
	send(t, conn, msg)
}

type eventData struct {
	event    int
	ch       string
	up, down uint64
}

func TestByteCount(t *testing.T) {
	out := make(chan eventData)
	handleSessionEvent := func(ch string, event int, up, down uint64) bool {
		out <- eventData{event, ch, up, down}
		return true
	}

	conn, ch := connect(t, handleSessionEvent, "")
	defer conn.Close()

	reader := bufio.NewReader(conn)

	receive(t, reader)
	receive(t, reader)

	sendByteCount(t, conn)

	data := <-out
	if data.event != SessionByteCount || data.ch != testChannel ||
		data.down != down || data.up != up {
		t.Fatalf("wrong up/down in agent mode")
	}

	assertNothingToReceive(t, conn, reader)

	exit(t, conn, ch)
}

func sendClientState(t *testing.T, conn net.Conn, connected bool) {
	var state string
	if connected {
		state = "CONNECTED"
	}
	msg := fmt.Sprintf("%s,%s", prefixState, state)
	send(t, conn, msg)
}

func TestClientSessionEvents(t *testing.T) {
	out := make(chan eventData)
	handleSessionEvent := func(ch string, event int, up, down uint64) bool {
		out <- eventData{event, ch, up, down}
		return true
	}

	conn, ch := connect(t, handleSessionEvent, testChannel)
	defer conn.Close()

	reader := bufio.NewReader(conn)

	receive(t, reader)
	receive(t, reader)
	receive(t, reader)

	// Should be ignored since it's not consireded
	// to be connected being in client mode.
	sendByteCountClient(t, conn)
	assertNothingToReceive(t, conn, reader)

	sendClientState(t, conn, true)

	data := <-out
	if data.event != SessionStarted {
		t.Fatalf("wrong event for started session in client mode")
	}

	sendByteCountClient(t, conn)

	data = <-out
	if data.event != SessionByteCount ||
		data.ch != testChannel ||
		data.down != down || data.up != up {
		t.Fatalf("wrong up/down in client mode")
	}

	sendClientState(t, conn, false)

	data = <-out
	if data.event != SessionStopped {
		t.Fatalf("wrong event for stopped session in client mode")
	}

	assertNothingToReceive(t, conn, reader)

	exit(t, conn, ch)
}

func TestKill(t *testing.T) {
	handleSessionEvent := func(ch string, event int, up, down uint64) bool {
		return false
	}

	conn, ch := connect(t, handleSessionEvent, "")
	defer conn.Close()

	reader := bufio.NewReader(conn)

	receive(t, reader)
	receive(t, reader)

	sendByteCount(t, conn)

	if str := receive(t, reader); str != "kill "+commonName {
		t.Fatalf("kill expected, but received: %s", str)
	}

	exit(t, conn, ch)
}

func TestMain(m *testing.M) {
	conf.FileLog = log.NewFileConfig()
	conf.VPNMonitor = NewConfig()
	util.ReadTestConfig(&conf)

	var err error
	logger, err = log.NewStderrLogger(conf.FileLog)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}
