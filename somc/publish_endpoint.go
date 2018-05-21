package somc

import (
	"encoding/json"
	"fmt"
)

const publishEndpointMethod = "connectionInfo"

type endpointParams struct {
	Channel  string `json:"stateChannel"`
	Endpoint []byte `json:"endpoint,omitempty"`
}

// PublishEndpoint publishes an endpoint for a state channel in SOMC.
func (c *Conn) PublishEndpoint(channel string, endpoint []byte) error {
	params := endpointParams{
		Channel:  channel,
		Endpoint: endpoint,
	}

	data, err := json.Marshal(&params)
	if err != nil {
		return fmt.Errorf("somc: could not marshal endpoint params: %v", err)
	}

	return c.request(publishEndpointMethod, data).err
}
