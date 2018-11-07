package somc

import (
	"encoding/json"

	"github.com/privatix/dappctrl/data"
)

const publishEndpointMethod = "connectionInfo"

// EndpointParams structure to store a Endpoint message in SOMC server.
type EndpointParams struct {
	Channel  data.Base64String `json:"stateChannel"`
	Endpoint []byte            `json:"endpoint,omitempty"`
}

// PublishEndpoint publishes an endpoint for a state channel in SOMC.
func (c *Conn) PublishEndpoint(channel data.Base64String, endpoint []byte) error {
	params := EndpointParams{
		Channel:  channel,
		Endpoint: endpoint,
	}

	data, err := json.Marshal(&params)
	if err != nil {
		return err
	}

	return c.request(publishEndpointMethod, data).err
}
