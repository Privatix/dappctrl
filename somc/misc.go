package somc

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
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

type reply struct {
	data []byte
	err  error
}

type write struct {
	msgType int
	data    []byte
	result  chan error
}

func (c *Conn) cancelPending(err error) {
	c.mtx.Lock()
	for k, v := range c.pending {
		v <- reply{nil, err}
		delete(c.pending, k)
	}
	c.mtx.Unlock()
	c.logger.Debug("canceled pending requests")
}

// isCloseNormalClosureErr checks a connection is terminated consistently.
func isCloseNormalClosureErr(err error) bool {
	if err2, ok := err.(*websocket.CloseError); ok {
		if err2.Code == websocket.CloseNormalClosure {
			return true
		}
	}
	return false
}

func (c *Conn) handleMessages() {
	logger := c.logger.Add("method", "handleMessages")

	for {
		var msg JSONRPCMessage
		if err := c.connection().ReadJSON(&msg); err != nil {
			if isCloseNormalClosureErr(err) {
				c.logger.Debug("normal close connection")
			} else {
				msg := fmt.Sprintf("failed to receive"+
					" message from SOMC, error: %s", err)
				logger.Warn(msg)
			}
			c.cancelPending(fmt.Errorf("response timeout"))
			time.Sleep(c.timeout)
			continue
		}
		c.handleMessage(&msg)
	}
}

func (c *Conn) handleMessage(m *JSONRPCMessage) {
	logger := c.logger.Add("method", "handleMessage")
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if len(m.Method) == 0 {
		if ch, ok := c.pending[m.ID]; ok {
			var err error
			if len(m.ErrorData) != 0 {
				logger.Warn("SOMC error: " +
					string(m.ErrorData))
				err = ErrInternal
			}
			ch <- reply{m.Result, err}
			delete(c.pending, m.ID)
		}
		return
	}
}

func (c *Conn) request(method string, params json.RawMessage) reply {
	ch := make(chan reply)

	logger := c.logger.Add("method", "request")

	c.mtx.Lock()

	c.id++
	body := JSONRPCMessage{
		Version: jsonRPCVersion,
		ID:      c.id,
		Method:  method,
		Params:  params,
	}
	data, _ := json.Marshal(body)
	result := make(chan error)

	request := &write{
		msgType: websocket.TextMessage,
		data:    data,
		result:  result,
	}

	c.write <- request
	if err := <-request.result; err != nil {
		c.mtx.Unlock()
		logger.Warn(err.Error())
		return reply{nil, err}
	}

	c.pending[body.ID] = ch

	c.mtx.Unlock()

	return <-ch
}

func (c *Conn) writeHandler() {
	for msg := range c.write {
		msg.result <- c.connection().WriteMessage(msg.msgType, msg.data)
	}
}

func (c *Conn) pingHandler() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill)

	logger := c.logger.Add("method", "pingHandler")

	for {
		select {
		// Sends PING message to server with current timestamp.
		case t := <-time.After(c.timeout / 3):
			result := make(chan error)
			request := &write{
				msgType: websocket.PingMessage,
				data:    []byte(fmt.Sprintf("%d", t.Unix())),
				result:  result,
			}

			c.write <- request
			if err := <-request.result; err != nil {
				logger.Warn(err.Error())
			}
		// If the application terminates, it closes the connection to
		// the server. See RFC 6455:
		// https://tools.ietf.org/html/rfc6455#section-7.1.1
		case <-interrupt:
			logger.Debug("interrupt")
			err := c.connection().WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(
					websocket.CloseNormalClosure, ""))
			if err != nil {
				msg := fmt.Sprintf(
					"write close message error: %s", err)
				logger.Warn(msg)
			}
			return
		}
	}
}
