package prepare

import (
	"path/filepath"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/svc/dappvpn/config"
	"github.com/privatix/dappctrl/svc/dappvpn/msg"
	"github.com/privatix/dappctrl/util/log"
)

// ClientConfig prepares configuration for Client.
// By the channel ID, finds a endpoint on a session server.
// Creates client configuration files for using a product.
func ClientConfig(logger log.Logger, channel string,
	adapterConfig *config.Config) error {
	args := sesssrv.EndpointMsgArgs{ChannelID: channel}

	var endpoint *data.Endpoint

	err := sesssrv.Post(adapterConfig.Server.Config, logger,
		adapterConfig.Server.Username, adapterConfig.Server.Password,
		sesssrv.PathEndpointMsg, args, &endpoint)
	if err != nil {
		logger.Add("channel", channel).Error(err.Error())
		return ErrGetEndpoint
	}

	save := func(str *string) string {
		if str != nil {
			return *str
		}
		return ""
	}

	target := filepath.Join(
		adapterConfig.OpenVPN.ConfigRoot, endpoint.Channel)

	err = msg.MakeFiles(logger, target,
		save(endpoint.ServiceEndpointAddress), save(endpoint.Username),
		save(endpoint.Password), endpoint.AdditionalParams,
		msg.SpecificOptions(adapterConfig.Monitor))
	if err != nil {
		return ErrMakeConfig
	}
	return nil
}
