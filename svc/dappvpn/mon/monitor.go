package mon

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/privatix/dappctrl/util/log"
)

// Config is a configuration for OpenVPN monitor.
type Config struct {
	Addr            string
	ByteCountPeriod uint // In seconds.
}

// NewConfig creates a default configuration for OpenVPN monitor.
func NewConfig() *Config {
	return &Config{
		Addr:            "localhost:7505",
		ByteCountPeriod: 5,
	}
}

type client struct {
	channel    string
	commonName string
}

// Monitor is an OpenVPN monitor for observation of consumed VPN traffic and
// for killing client VPN sessions.
type Monitor struct {
	conf            *Config
	logger          log.Logger
	handleSession   HandleSessionFunc
	channel         string // Client mode channel (empty in server mode).
	conn            net.Conn
	mtx             sync.Mutex // To guard writing.
	clients         map[uint]client
	clientConnected bool
}

// Session events.
const (
	SessionStarted   = iota // For client mode only.
	SessionStopped   = iota // For client mode only.
	SessionByteCount = iota
)

// HandleSessionFunc is a session event handler. If it returns false in server
// mode, then the monitor kills the corresponding session.
type HandleSessionFunc func(ch string, event int, up, down uint64) bool

// NewMonitor creates a new OpenVPN monitor.
func NewMonitor(conf *Config, logger log.Logger,
	handleSession HandleSessionFunc, channel string) *Monitor {
	return &Monitor{
		conf:          conf,
		logger:        logger.Add("type", "svc/dappvpn.Monitor"),
		handleSession: handleSession,
		channel:       channel,
	}
}

// Close immediately closes the monitor making MonitorTraffic() to return.
func (m *Monitor) Close() error {
	if m.conn != nil {
		return m.conn.Close()
	}
	return nil
}

// MonitorTraffic connects to OpenVPN management interfaces and starts
// monitoring VPN traffic.
func (m *Monitor) MonitorTraffic() error {
	logger := m.logger.Add("method", "MonitorTraffic")

	handleErr := func(err error) error {
		logger.Error(err.Error())
		return err
	}

	var err error
	if m.conn, err = net.Dial("tcp", m.conf.Addr); err != nil {
		return handleErr(err)
	}
	defer m.conn.Close()

	reader := bufio.NewReader(m.conn)

	if err := m.initConn(logger); err != nil {
		return handleErr(err)
	}

	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			return handleErr(err)
		}

		if err = m.processReply(logger, str); err != nil {
			return handleErr(err)
		}
	}
}

func (m *Monitor) write(cmd string) error {
	m.mtx.Lock()
	_, err := m.conn.Write([]byte(cmd + "\n"))
	m.mtx.Unlock()
	return err
}

func (m *Monitor) requestClients(logger log.Logger) error {
	logger.Info("requesting updated client list")
	return m.write("status 2")
}

func (m *Monitor) setByteCountPeriod() error {
	return m.write(fmt.Sprintf("bytecount %d", m.conf.ByteCountPeriod))
}

func (m *Monitor) killSession(cn string) error {
	return m.write(fmt.Sprintf("kill %s", cn))
}

func (m *Monitor) initConn(logger log.Logger) error {
	if err := m.setByteCountPeriod(); err != nil {
		return err
	}

	if len(m.channel) == 0 {
		if err := m.requestClients(logger); err != nil {
			return err
		}
	} else {
		if err := m.write("state on"); err != nil {
			return err
		}

		if err := m.write("hold release"); err != nil {
			return err
		}
	}

	return nil
}

const (
	prefixClientListHeader  = "HEADER,CLIENT_LIST,"
	prefixClientList        = "CLIENT_LIST,"
	prefixByteCount         = ">BYTECOUNT_CLI:"
	prefixByteCountClient   = ">BYTECOUNT:"
	prefixClientEstablished = ">CLIENT:ESTABLISHED,"
	prefixError             = "ERROR: "
	prefixState             = ">STATE:"
)

func (m *Monitor) processReply(logger log.Logger, s string) error {
	logger.Debug("openvpn raw: " + s)

	if strings.HasPrefix(s, prefixClientListHeader) {
		m.clients = make(map[uint]client)
		return nil
	}

	if strings.HasPrefix(s, prefixClientList) {
		return m.processClientList(logger, s[len(prefixClientList):])
	}

	if strings.HasPrefix(s, prefixByteCount) {
		return m.processByteCount(logger, s[len(prefixByteCount):])
	}

	if strings.HasPrefix(s, prefixByteCountClient) {
		return m.processByteCountClient(logger, s[len(prefixByteCountClient):])
	}

	if strings.HasPrefix(s, prefixClientEstablished) {
		return m.requestClients(logger)
	}

	if strings.HasPrefix(s, prefixState) {
		return m.processState(logger, s[len(prefixState):])
	}

	if strings.HasPrefix(s, prefixError) {
		logger.Error(s[len(prefixError):])
	}

	return nil
}

func split(s string) []string {
	return strings.Split(strings.TrimRight(s, "\r\n"), ",")
}

func (m *Monitor) processClientList(logger log.Logger, s string) error {
	sp := split(s)
	if len(sp) < 10 {
		return ErrServerOutdated
	}

	cid, err := strconv.ParseUint(sp[9], 10, 32)
	if err != nil {
		return err
	}

	m.clients[uint(cid)] = client{sp[8], sp[0]}

	logger.Add("cid", cid, "chan", sp[8], "cn", sp[0]).
		Info("openvpn client found")

	return nil
}

func (m *Monitor) processByteCount(logger log.Logger, s string) error {
	sp := split(s)

	cid, err := strconv.ParseUint(sp[0], 10, 32)
	if err != nil {
		return err
	}

	down, err := strconv.ParseUint(sp[1], 10, 64)
	if err != nil {
		return err
	}

	up, err := strconv.ParseUint(sp[2], 10, 64)
	if err != nil {
		return err
	}

	cl, ok := m.clients[uint(cid)]
	if !ok {
		return m.requestClients(logger)
	}

	logger.Add("chan", cl.channel, "up", up, "down", down).
		Info("openvpn byte count")

	go func() {
		if !m.handleSession(cl.channel, SessionByteCount, up, down) {
			m.killSession(cl.commonName)
		}
	}()

	return nil
}

func (m *Monitor) processByteCountClient(logger log.Logger, s string) error {
	if !m.clientConnected {
		return nil
	}

	sp := split(s)

	down, err := strconv.ParseUint(sp[0], 10, 64)
	if err != nil {
		return err
	}

	up, err := strconv.ParseUint(sp[1], 10, 64)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("openvpn byte count: up %d, down %d", up, down))

	go func() {
		m.handleSession(m.channel, SessionByteCount, up, down)
	}()

	return nil
}

func (m *Monitor) processState(logger log.Logger, s string) error {
	connected := split(s)[1] == "CONNECTED"

	if m.clientConnected && !connected {
		logger.Warn("disconnected from server")
		go func() {
			m.handleSession(m.channel, SessionStopped, 0, 0)
		}()
		m.clientConnected = false
	} else if !m.clientConnected && connected {
		logger.Warn("connected to server")
		go func() {
			m.handleSession(m.channel, SessionStarted, 0, 0)
		}()
		m.clientConnected = true
	}

	return nil
}
