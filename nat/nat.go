package nat

import (
	"context"
	"sync"
	"time"

	"github.com/privatix/dappctrl/util/log"
)

// Config is a configuration for service to traversal NAT.
type Config struct {
	CheckTimeout       int64  // In milliseconds.
	MapTimeout         int64  // In milliseconds.
	MapUpdateInterval  int64  // In milliseconds.
	Mechanism          string // upnp, pmp or any
	SoapRequestTimeout int64  // In milliseconds.
}

// NewConfig creates new configuration for service to traversal NAT.
func NewConfig() *Config {
	return &Config{
		CheckTimeout:       1000,
		MapTimeout:         1200000,
		MapUpdateInterval:  900000,
		Mechanism:          "any",
		SoapRequestTimeout: 3000,
	}
}

// Interface an implementation of nat.Interface can map local ports to ports
// accessible from the Internet.
type Interface interface {
	AddMapping(protocol string, extPort, intPort int,
		name string, lifetime time.Duration) error
	DeleteMapping(protocol string, extPort, intPort int) error
}

// Parse parses a NAT interface description.
func Parse(config *Config) (Interface, error) {
	switch config.Mechanism {
	case "any":
		return any(config), nil
	case "upnp":
		return uPnP(config), nil
	case "pmp":
		return pmp(), nil
	default:
		return nil, ErrBadMechanism
	}
}

// Map adds a port mapping on NAT interface and keeps it alive until interface
// is closed.
func Map(ctx context.Context, conf *Config, logger log.Logger, m Interface,
	protocol string, extPort, intPort int, name string) error {

	mapUpdateInterval := time.Duration(
		conf.MapUpdateInterval) * time.Millisecond

	mapTimeout := time.Duration(conf.MapTimeout) * time.Millisecond

	logger = logger.Add("proto", protocol, "extPort", extPort,
		"intPort", intPort, "mapUpdateInterval", mapUpdateInterval,
		"mapTimeout", mapTimeout, "name", name)

	if err := m.AddMapping(protocol, extPort, intPort,
		name, mapTimeout); err != nil {
		logger.Error(err.Error())
		return ErrAddMapping
	}
	logger.Info("mapped network port")
	go func() {
		timer := time.NewTimer(mapUpdateInterval)

		defer func() {
			timer.Stop()
			logger.Debug("deleting port mapping")
			m.DeleteMapping(protocol, extPort, intPort)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				logger.Debug("refreshing port mapping")
				if err := m.AddMapping(protocol, extPort,
					intPort, name, mapTimeout); err != nil {
					logger.Warn("couldn't add" +
						" port mapping, error: " +
						err.Error())
				}
				timer.Reset(mapUpdateInterval)
			}
		}
	}()
	return nil
}

func any(config *Config) Interface {
	return startDiscovery("UPnP or NAT-PMP", func() Interface {
		found := make(chan Interface, 2)
		go func() { found <- discoverUPnP(config) }()
		go func() { found <- discoverPMP() }()
		for i := 0; i < cap(found); i++ {
			if c := <-found; c != nil {
				return c
			}
		}
		return nil
	})
}

func uPnP(config *Config) Interface {
	return startDiscovery("UPnP",
		func() Interface { return discoverUPnP(config) })
}

func pmp() Interface {
	return startDiscovery("NAT-PMP", discoverPMP)
}

type discovery struct {
	what string
	once sync.Once
	do   func() Interface

	mu    sync.Mutex
	found Interface
}

func startDiscovery(what string, doit func() Interface) Interface {
	return &discovery{what: what, do: doit}
}

// AddMapping maps an external port to a local port for a specific
// service.
func (n *discovery) AddMapping(protocol string, extPort, intPort int,
	name string, lifetime time.Duration) error {
	if err := n.wait(); err != nil {
		return err
	}
	return n.found.AddMapping(protocol, extPort, intPort, name, lifetime)
}

// DeleteMapping removes the port mapping.
func (n *discovery) DeleteMapping(protocol string, extPort, intPort int) error {
	if err := n.wait(); err != nil {
		return err
	}
	return n.found.DeleteMapping(protocol, extPort, intPort)
}

func (n *discovery) wait() error {
	n.once.Do(func() {
		n.mu.Lock()
		n.found = n.do()
		n.mu.Unlock()
	})
	if n.found == nil {
		return ErrNoRouterDiscovered
	}
	return nil
}
