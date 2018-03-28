// +build !novpnmontest

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

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	vpnutil "github.com/privatix/dappctrl/vpn/util"
)

var conf struct {
	DB         *data.DBConfig
	Log        *util.LogConfig
	TestData   *data.TestData
	VPNMonitor *Config
}

var logger *util.Logger
var db *reform.DB

func connect(t *testing.T) (net.Conn, <-chan error) {
	lst, err := net.Listen("tcp", conf.VPNMonitor.Addr)
	if err != nil {
		t.Fatalf("failed to listen: %s", err)
	}
	defer lst.Close()

	ch := make(chan error)
	go func() {
		mon := NewMonitor(conf.VPNMonitor, logger, db)
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

func TestUnrecognizedOutput(t *testing.T) {
	conn, ch := connect(t)
	defer conn.Close()

	send(t, conn, "some text not recognized by the monitor")

	exit(t, conn, ch)
}

func TestCommError(t *testing.T) {
	conn, ch := connect(t)
	defer conn.Close()

	send(t, conn, prefixError+"some error")

	expectExit(t, ch, ErrComm)
}

func TestInitFlow(t *testing.T) {
	conn, ch := connect(t)
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

func TestOldOpenVPN(t *testing.T) {
	conn, ch := connect(t)
	defer conn.Close()

	send(t, conn, prefixClientListHeader)
	send(t, conn, prefixClientList+",,,,,,,,")

	expectExit(t, ch, ErrTooOldOpenVPN)
}

func createSession(t *testing.T) string {
	sess := data.Session{
		ID:      util.NewUUID(),
		Channel: conf.TestData.Channel,
		Started: time.Now(),
	}

	vsess := data.VPNSession{
		ID:         sess.ID,
		ServerIP:   "1.2.3.4",
		ClientIP:   "4.3.2.1",
		ClientPort: 1234,
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %s", err)
	}

	if err := db.Insert(&sess); err != nil {
		t.Fatalf("failed to insert session: %s", err)
	}

	if err := db.Insert(&vsess); err != nil {
		t.Fatalf("failed to insert vpn session: %s", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("failed to commit transaction: %s", err)
	}

	return sess.ID
}

const (
	cid        = 0
	up, down   = 1024, 2048
	commonName = "Common-Name"
)

func sendByteCount(t *testing.T, conn net.Conn) {
	send(t, conn, prefixClientListHeader)
	send(t, conn, fmt.Sprintf("%s%s,,,,,,,,%s,%d",
		prefixClientList, commonName, conf.TestData.Channel, cid))

	send(t, conn, fmt.Sprintf("%s%d,%d,%d", prefixByteCount, cid, down, up))
}

func TestByteCount(t *testing.T) {
	sid := createSession(t)

	conn, ch := connect(t)
	defer conn.Close()

	reader := bufio.NewReader(conn)

	receive(t, reader)
	receive(t, reader)

	sendByteCount(t, conn)

	assertNothingToReceive(t, conn, reader)

	exit(t, conn, ch)

	var vsess data.VPNSession
	if err := db.FindByPrimaryKeyTo(&vsess, sid); err != nil {
		t.Fatalf("failed to find vpn session: %s", err)
	}

	if vsess.Uploaded != up || vsess.Downloaded != down {
		t.Fatalf("wrong uploaded/downloaded values")
	}
}

func TestKill(t *testing.T) {
	vpnutil.SetChannelState(t, db,
		conf.TestData.Channel, data.ChannelClosing)
	defer vpnutil.SetChannelState(t, db,
		conf.TestData.Channel, data.ChannelOpen)

	createSession(t)

	conn, ch := connect(t)
	defer conn.Close()

	reader := bufio.NewReader(conn)

	receive(t, reader)
	receive(t, reader)

	sendByteCount(t, conn)

	if str := receive(t, reader); str != "kill "+commonName {
		t.Fatalf("unexpected kill command: %s", str)
	}

	exit(t, conn, ch)
}

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	conf.TestData = &data.TestData{}
	conf.VPNMonitor = NewConfig()
	util.ReadTestConfig(&conf)

	logger = util.NewTestLogger(conf.Log)

	db = data.NewTestDB(conf.DB, logger)
	defer data.CloseDB(db)

	os.Exit(m.Run())
}
