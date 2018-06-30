package somc

import (
	"encoding/json"
)

const getEndpointMethod = "getEndpoint"

// GetEndpoint requests SOMC to find endpoint by channel.
func (c *Conn) GetEndpoint(channel string) ([]byte, error) {
	params := EndpointParams{
		Channel: channel,
	}

	data, err := json.Marshal(&params)
	if err != nil {
		return nil, err
	}

	repl := c.request(getEndpointMethod, data)

	return repl.data, repl.err
}
