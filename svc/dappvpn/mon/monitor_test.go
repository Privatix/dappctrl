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
	Log            *util.LogConfig
	VPNMonitor     *Config
	VPNMonitorTest *testConfig
}

var logger *util.Logger

func connect(t *testing.T,
	handleByteCount HandleByteCountFunc) (net.Conn, <-chan error) {
	lst, err := net.Listen("tcp", conf.VPNMonitor.Addr)
	if err != nil {
		t.Fatalf("failed to listen: %s", err)
	}
	defer lst.Close()

	time.Sleep(time.Duration(conf.VPNMonitorTest.ServerStartupDelay) *
		time.Millisecond)

	ch := make(chan error)
	go func() {
		mon := NewMonitor(conf.VPNMonitor, logger, handleByteCount)
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

func ignoreByteCount(ch string, up, down uint64) bool {
	return true
}

func TestOldOpenVPN(t *testing.T) {
	conn, ch := connect(t, ignoreByteCount)
	defer conn.Close()

	send(t, conn, prefixClientListHeader)
	send(t, conn, prefixClientList+",,,,,,,,")

	expectExit(t, ch, ErrServerOutdated)
}

func TestInitFlow(t *testing.T) {
	conn, ch := connect(t, ignoreByteCount)
	defer conn.Close()

	reader := bufio.NewReader(conn)

	if str := receive(t, reader); str != "status 2" {
		t.Fatalf("unexpected status command: %s", str)
	}

	cmd := fmt.Sprintf("bytecount %d", conf.VPNMonitor.ByteCountPeriod)
	if str := receive(t, reader); str != cmd {
		t.Fatalf("unexpected bytecount command: %s", str)
	}

	exit(t, conn, ch)
}

const (
	cid         = 0
	up, down    = 1024, 2048
	commonName  = "Common-Name"
	testChannel = "Test-Channel"
)

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

func TestByteCount(t *testing.T) {
	type data struct {
		ch       string
		up, down uint64
	}

	out := make(chan data)
	handleByteCount := func(ch string, up, down uint64) bool {
		out <- data{ch, up, down}
		return true
	}

	conn, ch := connect(t, handleByteCount)
	defer conn.Close()

	reader := bufio.NewReader(conn)

	receive(t, reader)
	receive(t, reader)

	sendByteCount(t, conn)

	data2 := <-out
	if data2.ch != testChannel || data2.down != down || data2.up != up {
		t.Fatalf("wrong handler arguments for agent mode")
	}

	assertNothingToReceive(t, conn, reader)

	sendByteCountClient(t, conn)

	data2 = <-out
	if data2.ch != conf.VPNMonitor.Channel ||
		data2.down != down || data2.up != up {
		t.Fatalf("wrong handler arguments for client mode")
	}

	exit(t, conn, ch)
}

func TestKill(t *testing.T) {
	handleByteCount := func(ch string, up, down uint64) bool {
		return false
	}

	conn, ch := connect(t, handleByteCount)
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
	conf.Log = util.NewLogConfig()
	conf.VPNMonitor = NewConfig()
	util.ReadTestConfig(&conf)

	logger = util.NewTestLogger(conf.Log)

	os.Exit(m.Run())
}
