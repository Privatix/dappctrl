package somc

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

const jsonRPCVersion = "2.0"

// JSONRPCMessage is a format of the rpc message.
type JSONRPCMessage struct {
	Version   string          `json:"jsonrpc"`
	ID        uint32          `json:"id"`
	Method    string          `json:"method,omitempty"`
	Params    json.RawMessage `json:"params,omitempty"`
	Result    json.RawMessage `json:"result,omitempty"`
	ErrorData json.RawMessage `json:"error,omitempty"`
}

// IDString helper func returns message id as string.
func (m *JSONRPCMessage) IDString() string {
	return fmt.Sprint(m.ID)
}

// Error returns error if any.
func (m *JSONRPCMessage) Error() error {
	if len(m.ErrorData) == 0 {
		return nil
	}

	return fmt.Errorf("SOMC error: %+v", string(m.ErrorData))
}

type reply struct {
	data []byte
	err  error
}

func (c *Conn) cancelPending(err error) {
	c.mtx.Lock()
	for k, v := range c.pending {
		v <- reply{nil, err}
		delete(c.pending, k)
	}
	c.mtx.Unlock()
}

func (c *Conn) handleMessages() {
	for !c.exit {
		var err error
		for {
			var msg JSONRPCMessage
			if err = c.conn.ReadJSON(&msg); err != nil {
				break
			}

			c.handleMessage(&msg)
		}

		c.cancelPending(err)

		if c.exit {
			break
		}

		c.logger.Error("failed to receive SOMC response: %s", err)

		// Make sure following requests will fail.
		c.conn.Close()

		for !c.exit {
			conn, _, err :=
				websocket.DefaultDialer.Dial(c.conf.URL, nil)
			if err == nil {
				c.conn = conn
				break
			}

			c.logger.Error("failed to reconnect to SOMC: %s", err)
			c.cancelPending(err)

			time.Sleep(time.Millisecond *
				time.Duration(c.conf.ReconnPeriod))
		}
	}
}

func (c *Conn) handleMessage(m *JSONRPCMessage) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if len(m.Method) == 0 {
		if ch, ok := c.pending[m.IDString()]; ok {
			ch <- reply{m.Result, m.Error()}
			delete(c.pending, m.IDString())
		}
		return
	}

	// Endpoint message has arrived.

	var params endpointParams
	if err := json.Unmarshal(m.Params, &params); err != nil {
		c.logger.Error("failed to unmarshal SOMC params: %s", err)
		return
	}

	if ch, ok := c.pending[params.Channel]; ok {
		ch <- reply{params.Endpoint, nil}
		delete(c.pending, params.Channel)
	}
}

func (c *Conn) request(method string, params json.RawMessage) reply {
	ch := make(chan reply)

	c.mtx.Lock()

	c.id++
	msg := JSONRPCMessage{
		Version: jsonRPCVersion,
		ID:      c.id,
		Method:  method,
		Params:  params,
	}

	if err := c.conn.WriteJSON(&msg); err != nil {
		c.mtx.Unlock()
		return reply{nil, err}
	}

	c.pending[msg.IDString()] = ch

	c.mtx.Unlock()

	return <-ch
}
