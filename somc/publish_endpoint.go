package somc

import (
	"encoding/json"
	"fmt"
)

const publishEndpointMethod = "connectionInfo"

// EndpointParams structure to store a Endpoint message in SOMC server.
type EndpointParams struct {
	Channel  string `json:"stateChannel"`
	Endpoint []byte `json:"endpoint,omitempty"`
}

// PublishEndpoint publishes an endpoint for a state channel in SOMC.
func (c *Conn) PublishEndpoint(channel string, endpoint []byte) error {
	params := EndpointParams{
		Channel:  channel,
		Endpoint: endpoint,
	}

	data, err := json.Marshal(&params)
	if err != nil {
		return fmt.Errorf("somc: could not marshal endpoint params: %v", err)
	}

	return c.request(publishEndpointMethod, data).err
}
