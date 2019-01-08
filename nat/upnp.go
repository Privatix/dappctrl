package nat

import (
	"net"
	"strings"
	"time"

	"github.com/huin/goupnp"
	"github.com/huin/goupnp/dcps/internetgateway1"
	"github.com/huin/goupnp/dcps/internetgateway2"
)

type upnp struct {
	dev     *goupnp.RootDevice
	service string
	client  upnpClient
}

// upnpClient an implementation of UPnP client.
type upnpClient interface {
	GetExternalIPAddress() (string, error)
	AddPortMapping(string, uint16, string,
		uint16, string, bool, string, uint32) error
	DeletePortMapping(string, uint16, string) error
	GetNATRSIPStatus() (sip bool, nat bool, err error)
}

// AddMapping maps an external port to a local port for a specific
// service to UPnP interface.
func (n *upnp) AddMapping(protocol string, extPort, intPort int,
	desc string, lifetime time.Duration) error {
	ip, err := n.internalAddress()
	if err != nil {
		return nil
	}
	protocol = strings.ToUpper(protocol)
	lifetimeS := uint32(lifetime / time.Second)
	n.DeleteMapping(protocol, extPort, intPort)
	return n.client.AddPortMapping("", uint16(extPort),
		protocol, uint16(intPort), ip.String(), true, desc, lifetimeS)
}

func (n *upnp) internalAddress() (net.IP, error) {
	devAddr, err := net.ResolveUDPAddr("udp4", n.dev.URLBase.Host)
	if err != nil {
		return nil, err
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}

		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			x, ok := addr.(*net.IPNet)
			if ok && x.Contains(devAddr.IP) {
				return x.IP, nil
			}
		}
	}
	return nil, ErrLocalAddressNotFound
}

// DeleteMapping removes the port mapping to UPnP interface.
func (n *upnp) DeleteMapping(protocol string, extPort, intPort int) error {
	return n.client.DeletePortMapping(
		"", uint16(extPort), strings.ToUpper(protocol))
}

func discoverUPnP(config *Config) Interface {
	found := make(chan *upnp, 2)
	// IGDv1
	go discover(config, found, internetgateway1.URN_WANConnectionDevice_1,
		func(dev *goupnp.RootDevice, sc goupnp.ServiceClient) *upnp {
			switch sc.Service.ServiceType {
			case internetgateway1.URN_WANIPConnection_1:
				return &upnp{dev, "IGDv1-IP1",
					&internetgateway1.WANIPConnection1{ServiceClient: sc}}
			case internetgateway1.URN_WANPPPConnection_1:
				return &upnp{dev, "IGDv1-PPP1",
					&internetgateway1.WANPPPConnection1{ServiceClient: sc}}
			}
			return nil
		})
	// IGDv2
	go discover(config, found, internetgateway2.URN_WANConnectionDevice_2,
		func(dev *goupnp.RootDevice, sc goupnp.ServiceClient) *upnp {
			switch sc.Service.ServiceType {
			case internetgateway2.URN_WANIPConnection_1:
				return &upnp{dev, "IGDv2-IP1",
					&internetgateway2.WANIPConnection1{ServiceClient: sc}}
			case internetgateway2.URN_WANIPConnection_2:
				return &upnp{dev, "IGDv2-IP2",
					&internetgateway2.WANIPConnection2{ServiceClient: sc}}
			case internetgateway2.URN_WANPPPConnection_1:
				return &upnp{dev, "IGDv2-PPP1",
					&internetgateway2.WANPPPConnection1{ServiceClient: sc}}
			}
			return nil
		})
	for i := 0; i < cap(found); i++ {
		if c := <-found; c != nil {
			return c
		}
	}
	return nil
}

func discover(config *Config, out chan<- *upnp, target string,
	matcher func(*goupnp.RootDevice, goupnp.ServiceClient) *upnp) {
	devs, err := goupnp.DiscoverDevices(target)
	if err != nil {
		out <- nil
		return
	}
	found := false
	for i := 0; i < len(devs) && !found; i++ {
		if devs[i].Root == nil {
			continue
		}
		devs[i].Root.Device.VisitServices(func(service *goupnp.Service) {
			if found {
				return
			}
			sc := goupnp.ServiceClient{
				SOAPClient: service.NewSOAPClient(),
				RootDevice: devs[i].Root,
				Location:   devs[i].Location,
				Service:    service,
			}
			soapRequestTimeout := time.Duration(
				config.SoapRequestTimeout) * time.Millisecond
			sc.SOAPClient.HTTPClient.Timeout = soapRequestTimeout
			upnp := matcher(devs[i].Root, sc)
			if upnp == nil {
				return
			}
			_, nat, err := upnp.client.GetNATRSIPStatus()
			if err != nil || !nat {
				return
			}
			out <- upnp
			found = true
		})
	}
	if !found {
		out <- nil
	}
}
