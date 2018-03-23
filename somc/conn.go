package somc

import (
	"sync"

	"github.com/gorilla/websocket"

	"github.com/privatix/dappctrl/util"
)

// Config is a configuration for SOMC connection.
type Config struct {
	ReconnPeriod int // In milliseconds.
	URL          string
}

// NewConfig creates a default configuration for SOMC connection.
func NewConfig() *Config {
	return &Config{
		ReconnPeriod: 5000,
		URL:          "ws://localhost:8080",
	}
}

// Conn is a websocket connection to SOMC.
type Conn struct {
	conf    *Config
	logger  *util.Logger
	conn    *websocket.Conn
	pending map[string]chan reply
	mtx     sync.Mutex // Mostly to guard the pending map.
	exit    bool
	id      uint32
}

// NewConn creates and starts a new SOMC connection.
func NewConn(conf *Config, logger *util.Logger) (*Conn, error) {
	c, _, err := websocket.DefaultDialer.Dial(conf.URL, nil)
	if err != nil {
		return nil, err
	}

	conn := &Conn{
		conf:    conf,
		logger:  logger,
		conn:    c,
		pending: make(map[string]chan reply),
	}

	go conn.handleMessages()

	return conn, nil
}

// Close closes a given SOMC connection.
func (c *Conn) Close() error {
	c.exit = true
	return c.conn.Close()
}
