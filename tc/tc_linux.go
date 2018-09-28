package tc

import (
	"fmt"
	"hash/crc32"
	"net"
	"strings"

	"github.com/privatix/dappctrl/util/log"
)

// See http://tldp.org/HOWTO/Traffic-Control-HOWTO/ as reference.

// Config is a traffic control configuration.
type Config struct {
	IpPath       string
	IptablesPath string
	TcPath       string
	TunRateRatio float32 // Bandwidth overhead ratio for tunnel interface.
}

// NewConfig creates a default configuration.
func NewConfig() *Config {
	return &Config{
		IpPath:       "/sbin/ip",
		IptablesPath: "/sbin/iptables",
		TcPath:       "/sbin/tc",
		TunRateRatio: 1.25,
	}
}

func (tc *TrafficControl) findDefaultIface(logger log.Logger) (string, error) {
	out, err := tc.run(logger, tc.conf.IpPath, "route")
	if err != nil {
		return "", err
	}

	for _, v := range strings.Split(out, "\n") {
		if !strings.HasPrefix(v, "default ") {
			continue
		}

		tabs := strings.Split(v, " ")
		if len(tabs) < 5 || tabs[3] != "dev" {
			break
		}

		return tabs[4], nil
	}

	return "", ErrFailedToFindDefaultIface
}

func classID(ip net.IP, down bool) string {
	hash := uint16(crc32.ChecksumIEEE(ip))
	if down {
		hash++
	}
	return fmt.Sprintf("1:%x", hash)
}

func withMask(ip net.IP) string {
	bytes := 16
	if ip.To4() != nil {
		bytes = 4
	}
	return fmt.Sprintf("%s/%d", ip, bytes*8)
}

func (tc *TrafficControl) classify(logger log.Logger,
	cid string, ip net.IP, down, disable bool) error {
	act := "-A"
	if disable {
		act = "-D"
	}

	dir := "-s"
	if down {
		dir = "-d"
	}

	_, err := tc.run(logger, tc.conf.IptablesPath,
		"-t", "mangle", act, "POSTROUTING", dir,
		withMask(ip), "-j", "CLASSIFY", "--set-class", cid)

	return err
}

func (tc *TrafficControl) setRateClass(logger log.Logger,
	iface, cid string, ip net.IP, down bool, mbits float32) error {
	rate := fmt.Sprintf("%fMbit", mbits*tc.conf.TunRateRatio)

	_, err := tc.run(logger, tc.conf.TcPath,
		"class", "add", "dev", iface, "parent", "1:",
		"classid", cid, "htb", "rate", rate, "ceil", rate)
	if err != nil {
		return err
	}

	return tc.classify(logger, cid, ip, down, false)
}

func (tc *TrafficControl) unsetRateClass(logger log.Logger,
	iface, cid string, ip net.IP, down bool) error {
	if err := tc.classify(logger, cid, ip, down, true); err != nil {
		return err
	}

	_, err := tc.run(logger, tc.conf.TcPath,
		"qdisc", "del", "dev", iface, "root", "handle", cid, "htb")

	return err
}

// SetRateLimit sets a rate limit for a given client IP address on a given
// network interface.
func (tc *TrafficControl) SetRateLimit(
	clientIP string, upMbits, downMbits float32) error {
	logger := tc.logger.Add("method", "SetRateLimit",
		"clientIp", clientIP, "up", upMbits, "down", downMbits)

	iface, err := tc.findDefaultIface(logger)
	if err != nil {
		return err
	}

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

	if err := tc.setRateClass(logger, iface,
		classID(ip, false), ip, false, upMbits); err != nil {
		return err
	}

	if err := tc.setRateClass(logger, iface,
		classID(ip, true), ip, true, downMbits); err != nil {
		return err
	}

	return nil
}

// UnsetRateLimit removes a rate limit for a given client IP address on a given
// network interface.
func (tc *TrafficControl) UnsetRateLimit(clientIP string) error {
	logger := tc.logger.Add(
		"method", "UnsetRateLimit", "clientIp", clientIP)

	iface, err := tc.findDefaultIface(logger)
	if err != nil {
		return err
	}

	ip := net.ParseIP(clientIP)
	if ip == nil {
		return ErrBadClientIP
	}

	if err := tc.unsetRateClass(logger, iface,
		classID(ip, false), ip, false); err != nil {
		return err
	}

	if err := tc.unsetRateClass(logger, iface,
		classID(ip, true), ip, true); err != nil {
		return err
	}

	return nil
}
