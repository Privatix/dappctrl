package prepare

import (
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/svc/dappvpn/config"
	"github.com/privatix/dappctrl/svc/dappvpn/msg"
)

// ClientConfig prepares configuration for Client.
// By the channel ID, finds a endpoint on a session server.
// Creates client configuration files for using a product.
func ClientConfig(channel string, adapterConfig *config.Config) error {
	args := sesssrv.EndpointMsgArgs{ChannelID: channel}

	var endpoint *data.Endpoint

	err := sesssrv.Post(adapterConfig.Server.Config,
		adapterConfig.Server.Username, adapterConfig.Server.Password,
		sesssrv.PathEndpointMsg, args, &endpoint)
	if err != nil {
		return errors.Wrap(err, "failed to get endpoint")
	}

	save := func(str *string) string {
		if str != nil {
			return *str
		}
		return ""
	}

	target := filepath.Join(
		adapterConfig.OpenVPN.ConfigRoot, endpoint.Channel)

	err = msg.MakeFiles(target,
		save(endpoint.ServiceEndpointAddress), save(endpoint.Username),
		save(endpoint.Password), endpoint.AdditionalParams,
		msg.SpecificOptions(adapterConfig.Monitor))
	if err != nil {
		return errors.Wrap(err, "failed to make client"+
			" configuration files")
	}
	return err
}
