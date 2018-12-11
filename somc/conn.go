package somc

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/privatix/dappctrl/util/log"
)

// Config is a configuration for SOMC connection.
type Config struct {
	CheckTimeout int // In milliseconds.
	URL          string
}

// NewConfig creates a default configuration for SOMC connection.
func NewConfig() *Config {
	return &Config{
		CheckTimeout: 30000,
		URL:          "ws://localhost:8080",
	}
}

// Conn is a websocket connection to SOMC.
type Conn struct {
	conf        *Config
	logger      log.Logger
	id          uint32
	pongChannel chan interface{}
	timeout     time.Duration
	write       chan *write

	mtx     sync.Mutex
	pending map[uint32]chan reply

	mtx2 sync.Mutex
	conn *websocket.Conn
}

// NewConn creates and starts a new SOMC connection.
func NewConn(conf *Config, logger log.Logger) (*Conn, error) {
	timeout := time.Duration(conf.CheckTimeout) * time.Millisecond

	conn := &Conn{
		conf:        conf,
		logger:      logger.Add("type", "somc.Conn"),
		pending:     make(map[uint32]chan reply),
		pongChannel: make(chan interface{}),
		timeout:     timeout,
		write:       make(chan *write),
	}

	if err := conn.connect(); err != nil {
		return nil, err
	}

	go conn.handleMessages()
	go conn.pingHandler()
	go conn.writeHandler()
	go conn.connectionControl()

	return conn, nil
}

// Close closes a given SOMC connection.
func (c *Conn) Close() error {
	c.mtx2.Lock()
	defer c.mtx2.Unlock()
	return c.conn.Close()
}

func (c *Conn) connection() *websocket.Conn {
	c.mtx2.Lock()
	defer c.mtx2.Unlock()
	return c.conn
}

// pongHandler is function to receives PONG messages from websocket server.
// Sends pong message to state channel for further processing.
func (c *Conn) pongHandler(pong string) error {
	c.pongChannel <- pong
	return nil
}

// connectionControl receives PONG messages from state channel. If PONG message
// is not received within a certain period of time, initializes the reconnection
// to a websocket server.
func (c *Conn) connectionControl() {
	logger := c.logger.Add("method", "connectionControl",
		"timeout", c.timeout.String())

	connected := true

	for {
		select {
		case msg := <-c.pongChannel:
			if !connected {
				connected = true
				logger.Warn("SOMC communication restored")
				continue
			}
			logger.Add("message",
				msg).Debug("SOMC communication checked")
		case <-time.After(c.timeout):
			logger.Warn("reconnecting to SOMC")
			connected = false
			// Clears map for pending requests.
			c.cancelPending(fmt.Errorf("response timeout"))
			// Closes connection to server.
			c.connection().Close()
			// Trying to connect to the server.
			if err := c.connect(); err != nil {
				logger.Warn(fmt.Sprintf("failed to"+
					" reconnect to SOMC: %s", err))
			}
		}
	}
}

func (c *Conn) connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.conf.URL, nil)
	if err != nil {
		return err
	}
	conn.SetPongHandler(c.pongHandler)

	c.mtx2.Lock()
	defer c.mtx2.Unlock()
	c.conn = conn
	return nil
}
