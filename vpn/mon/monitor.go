package mon

import (
	"errors"
	"fmt"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	vpnutil "github.com/privatix/dappctrl/vpn/util"
	"github.com/ziutek/telnet"
	"gopkg.in/reform.v1"
	"strconv"
	"strings"
)

// Config is a configuration for OpenVPN monitor.
type Config struct {
	Addr            string
	ByteCountPeriod uint
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
	conf    *Config
	logger  *util.Logger
	db      *reform.DB
	conn    *telnet.Conn
	clients map[uint]client
}

// NewMonitor creates a new OpenVPN monitor.
func NewMonitor(conf *Config, logger *util.Logger, db *reform.DB) *Monitor {
	return &Monitor{conf, logger, db, nil, nil}
}

// MonitorTraffic connects to OpenVPN management interfaces and starts
// monitoring VPN traffic.
func (m *Monitor) MonitorTraffic() error {
	var err error
	if m.conn, err = telnet.Dial("tcp", m.conf.Addr); err != nil {
		return err
	}
	defer m.conn.Close()

	if err := m.requestClients(); err != nil {
		return err
	}

	cmd := fmt.Sprintf("bytecount %d\n", m.conf.ByteCountPeriod)
	if _, err := m.conn.Write([]byte(cmd)); err != nil {
		return err
	}

	for {
		str, err := m.conn.ReadString('\n')
		if err != nil {
			return err
		}

		if err = m.processReply(str); err != nil {
			return err
		}
	}
}

func (m *Monitor) requestClients() error {
	if _, err := m.conn.Write([]byte("status 2\n")); err != nil {
		return err
	}
	return nil
}

const (
	prefixClientListHeader  = "HEADER,CLIENT_LIST,"
	prefixClientList        = "CLIENT_LIST,"
	prefixByteCount         = ">BYTECOUNT_CLI:"
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

	if strings.HasPrefix(s, prefixClientEstablished) {
		return m.requestClients()
	}

	if strings.HasPrefix(s, prefixError) {
		m.logger.Error("openvpn error: %s", s[len(prefixError):])
		return errors.New("openvpn communication error")
	}

	return nil
}

func (m *Monitor) processClientList(s string) error {
	sp := strings.Split(s, ",")
	if len(sp) < 10 {
		return errors.New("no cid returned for 'status 2' " +
			"(too old openvpn version)")
	}

	cid, err := strconv.Atoi(sp[9])
	if err != nil {
		return err
	}

	m.clients[uint(cid)] = client{sp[8], sp[0]}
	m.logger.Info("openvpn client found: cid %d, chan %s, cn %s",
		cid, sp[8], sp[0])

	return nil
}

func (m *Monitor) processByteCount(s string) error {
	sp := strings.Split(strings.TrimRight(s, "\r\n"), ",")

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
		m.logger.Warn("unknown openvpn cid: %d", cid)
		return m.requestClients()
	}

	err = m.updateBytes(&cl, up, down)
	if err != nil {
		return err
	}

	return nil
}

func (m *Monitor) updateBytes(cl *client, up, down uint64) error {
	sid, err := vpnutil.FindCurrentSession(m.db, cl.channel)
	if err != nil {
		m.logger.Error("no session to update bytes for channel %s: %s",
			cl.channel, err)
		return err
	}

	var sess data.VPNSession
	if err := m.db.FindByPrimaryKeyTo(&sess, sid); err != nil {
		m.logger.Error("failed to find session %s: %s", sid, err)
		return err
	}

	sess.Uploaded = up
	sess.Downloaded = down

	if err := m.db.Save(&sess); err != nil {
		m.logger.Error("failed to save session %s: %s", sid, err)
		return err
	}

	positive, err := vpnutil.HasPositiveBalance(m.db, cl.channel)
	if err != nil {
		m.logger.Error("failed to check for positive balance: %s", err)
		return err
	}

	if !positive {
		m.logger.Info("killing client session %s", sid)

		cmd := fmt.Sprintf("kill %s\n", cl.commonName)
		if _, err := m.conn.Write([]byte(cmd)); err != nil {
			return err
		}
	}

	return nil
}
