package tc

import (
	"fmt"
	"hash/crc32"
	"net"
	"strings"
)

// Config is a traffic control configuration.
type Config struct {
	TcPath       string
	IptablesPath string
}

// NewConfig creates a default configuration.
func NewConfig() *Config {
	return &Config{
		TcPath:       "/sbin/tc",
		IptablesPath: "/sbin/iptables",
	}
}

// See http://tldp.org/HOWTO/Traffic-Control-HOWTO/ as reference.

// SetRateLimit sets a rate limit for a given client IP address on a given
// network interface.
func (tc *TrafficControl) SetRateLimit(
	iface, clientIP string, upMbps, downMbps int) error {
	logger := tc.logger.Add("method", "SetRateLimit", "iface", iface,
		"clientIp", clientIP, "up", upMbps, "down", downMbps)

	ip := net.ParseIP(clientIP)
	if ip == nil {
		return ErrBadClientIP
	}

	out, err := tc.run(logger, tc.conf.TcPath,
		"-s", "-d", "qdisc", "show", "dev", iface)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(out, "qdisc htb 1: root ") {
		logger.Info("setting a root htb discipline")
		_, err := tc.run(logger, tc.conf.TcPath, "qdisc",
			"add", "dev", iface, "root", "handle", "1:", "htb")
		if err != nil {
			return err
		}
	}

	cid := classID(ip)
	rate := fmt.Sprintf("%dMbit", downMbps)
	_, err = tc.run(logger, tc.conf.TcPath,
		"class", "add", "dev", iface, "parent", "1:",
		"classid", cid, "htb", "rate", rate, "ceil", rate)
	if err != nil {
		return err
	}

	_, err = tc.run(logger, tc.conf.IptablesPath,
		"-t", "mangle", "-A", "POSTROUTING", "-o", iface,
		"-d", withMask(ip), "-j", "CLASSIFY", "--set-class", cid)

	return err
}

// UnsetRateLimit removes a rate limit for a given client IP address on a given
// network interface.
func (tc *TrafficControl) UnsetRateLimit(iface, clientIP string) error {
	logger := tc.logger.Add("method", "UnsetRateLimit",
		"iface", iface, "clientIp", clientIP)

	ip := net.ParseIP(clientIP)
	if ip == nil {
		return ErrBadClientIP
	}

	cid := classID(ip)
	_, err := tc.run(logger, tc.conf.IptablesPath,
		"-t", "mangle", "-D", "POSTROUTING", "-o", iface,
		"-d", withMask(ip), "-j", "CLASSIFY", "--set-class", cid)
	if err != nil {
		return err
	}

	_, err = tc.run(logger, tc.conf.TcPath,
		"qdisc", "del", "dev", iface, "root", "handle", cid, "htb")
	if err != nil {
		return err
	}

	return nil
}

func classID(ip net.IP) string {
	return fmt.Sprintf("1:%x", uint16(crc32.ChecksumIEEE(ip)))
}

func withMask(ip net.IP) string {
	return fmt.Sprintf("%s/%d", ip, len(ip))
}
