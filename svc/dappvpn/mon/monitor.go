package mon

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/privatix/dappctrl/util"
)

// Config is a configuration for OpenVPN monitor.
type Config struct {
	Addr            string
	ByteCountPeriod uint   // In seconds.
	Channel         string // Client mode channel.
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
	logger          *util.Logger
	handleByteCount HandleByteCountFunc
	conn            net.Conn
	mtx             sync.Mutex // To guard writing.
	clients         map[uint]client
}

// HandleByteCountFunc is a byte count handler. If it returns false, then the
// monitor kills the corresponding session.
type HandleByteCountFunc func(ch string, up, down uint64) bool

// NewMonitor creates a new OpenVPN monitor.
func NewMonitor(conf *Config, logger *util.Logger,
	handleByteCount HandleByteCountFunc) *Monitor {
	return &Monitor{
		conf:            conf,
		logger:          logger,
		handleByteCount: handleByteCount,
	}
}

// Close immediately closes the monitor making MonitorTraffic() to return.
func (m *Monitor) Close() error {
	if m.conn != nil {
		return m.conn.Close()
	}
	return nil
}

// Monitor errors.
var (
	ErrServerOutdated = errors.New("server outdated")
)

// MonitorTraffic connects to OpenVPN management interfaces and starts
// monitoring VPN traffic.
func (m *Monitor) MonitorTraffic() error {
	var err error
	if m.conn, err = net.Dial("tcp", m.conf.Addr); err != nil {
		return err
	}
	defer m.conn.Close()

	reader := bufio.NewReader(m.conn)

	if err := m.requestClients(); err != nil {
		return err
	}

	if err := m.setByteCountPeriod(); err != nil {
		return err
	}

	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		if err = m.processReply(str); err != nil {
			return err
		}
	}
}

func (m *Monitor) write(cmd string) error {
	m.mtx.Lock()
	_, err := m.conn.Write([]byte(cmd + "\n"))
	m.mtx.Unlock()
	return err
}

func (m *Monitor) requestClients() error {
	m.logger.Info("requesting updated client list")
	return m.write("status 2")
}

func (m *Monitor) setByteCountPeriod() error {
	return m.write(fmt.Sprintf("bytecount %d", m.conf.ByteCountPeriod))
}

func (m *Monitor) killSession(cn string) error {
	return m.write(fmt.Sprintf("kill %s", cn))
}

const (
	prefixClientListHeader  = "HEADER,CLIENT_LIST,"
	prefixClientList        = "CLIENT_LIST,"
	prefixByteCount         = ">BYTECOUNT_CLI:"
	prefixByteCountClient   = ">BYTECOUNT:"
	prefixClientEstablished = ">CLIENT:ESTABLISHED,"
	prefixError             = "ERROR: "
)

func (m *Monitor) processReply(s string) error {
	m.logger.Debug("openvpn raw: %s", s)

	if strings.HasPrefix(s, prefixClientListHeader) {
		m.clients = make(map[uint]client)
		return nil
	}

	if strings.HasPrefix(s, prefixClientList) {
		return m.processClientList(s[len(prefixClientList):])
	}

	if strings.HasPrefix(s, prefixByteCount) {
		return m.processByteCount(s[len(prefixByteCount):])
	}

	if strings.HasPrefix(s, prefixByteCountClient) {
		return m.processByteCountClient(s[len(prefixByteCountClient):])
	}

	if strings.HasPrefix(s, prefixClientEstablished) {
		return m.requestClients()
	}

	if strings.HasPrefix(s, prefixError) {
		m.logger.Error("openvpn error: %s", s[len(prefixError):])
	}

	return nil
}

func split(s string) []string {
	return strings.Split(strings.TrimRight(s, "\r\n"), ",")
}

func (m *Monitor) processClientList(s string) error {
	sp := split(s)
	if len(sp) < 10 {
		return ErrServerOutdated
	}

	cid, err := strconv.ParseUint(sp[9], 10, 32)
	if err != nil {
		return err
	}

	m.clients[uint(cid)] = client{sp[8], sp[0]}
	m.logger.Info("openvpn client found: cid %d, chan %s, cn %s",
		cid, sp[8], sp[0])

	return nil
}

func (m *Monitor) processByteCount(s string) error {
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
		return m.requestClients()
	}

	m.logger.Info("openvpn byte count for chan %s: up %d, down %d",
		cl.channel, up, down)

	go func() {
		if !m.handleByteCount(cl.channel, up, down) {
			m.killSession(cl.commonName)
		}
	}()

	return nil
}

func (m *Monitor) processByteCountClient(s string) error {
	sp := split(s)

	down, err := strconv.ParseUint(sp[0], 10, 64)
	if err != nil {
		return err
	}

	up, err := strconv.ParseUint(sp[1], 10, 64)
	if err != nil {
		return err
	}

	m.logger.Info("openvpn byte count: up %d, down %d", up, down)

	go func() {
		m.handleByteCount(m.conf.Channel, up, down)
	}()

	return nil
}
