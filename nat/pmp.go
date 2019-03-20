package nat

import (
	"net"
	"strings"
	"time"

	"github.com/jackpal/go-nat-pmp"
)

var (
	_, private24BitBlock, _ = net.ParseCIDR("10.0.0.0/8")
	_, private20BitBlock, _ = net.ParseCIDR("172.16.0.0/12")
	_, private16BitBlock, _ = net.ParseCIDR("192.168.0.0/16")
)

type npmp struct {
	gw net.IP
	c  *natpmp.Client
}

// AddMapping maps an external port to a local port for a specific
// service to NAT-PMP interface.
func (n *npmp) AddMapping(protocol string, extPort, intPort int,
	name string, lifetime time.Duration) error {
	if lifetime <= 0 {
		return ErrTooShortLifetime
	}
	_, err := n.c.AddPortMapping(strings.ToLower(protocol),
		intPort, extPort, int(lifetime/time.Second))
	return err
}

// DeleteMapping removes the port mapping to NAT-PMP interface.
func (n *npmp) DeleteMapping(protocol string, extPort, intPort int) (err error) {
	_, err = n.c.AddPortMapping(strings.ToLower(protocol), intPort, 0, 0)
	return err
}

func discoverPMP() Interface {
	gws := potentialGateways()
	found := make(chan *npmp, len(gws))
	for i := range gws {
		gw := gws[i]
		go func() {
			c := natpmp.NewClient(gw)
			if _, err := c.GetExternalAddress(); err != nil {
				found <- nil
			} else {
				found <- &npmp{gw, c}
			}
		}()
	}
	timeout := time.NewTimer(1 * time.Second)
	defer timeout.Stop()
	for range gws {
		select {
		case c := <-found:
			if c != nil {
				return c
			}
		case <-timeout.C:
			return nil
		}
	}
	return nil
}

func potentialGateways() (gws []net.IP) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, iface := range ifaces {
		ifAddrs, err := iface.Addrs()
		if err != nil {
			return gws
		}
		for _, addr := range ifAddrs {
			x, ok := addr.(*net.IPNet)
			if ok {
				if private24BitBlock.Contains(x.IP) ||
					private20BitBlock.Contains(x.IP) ||
					private16BitBlock.Contains(x.IP) {
					ip := x.IP.Mask(x.Mask).To4()
					if ip != nil {
						ip[3] = ip[3] | 0x01
						gws = append(gws, ip)
					}
				}
			}
		}
	}
	return gws
}
