package somcserver

import (
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
)

// Endpoint returns endpoint msg for a channel with given key.
func (h *Handler) Endpoint(key data.Base64String) (*data.Base64String, error) {
	logger := h.logger.Add("type", "agent/tor-somc.Handler")

	channelsStructs, err := h.db.SelectAllFrom(data.ChannelTable, "")
	if err != nil {
		h.logger.Error(err.Error())
		return nil, ErrInternal
	}

	for _, chanStruct := range channelsStructs {
		channel := chanStruct.(*data.Channel)
		channelKey, err := h.channelKey(logger, channel)
		if err != nil {
			h.logger.Error(err.Error())
			return nil, ErrInternal
		}
		if channelKey == key {
			endpoint, err := h.endpointByChannelID(logger, channel.ID)
			if err != nil {
				return nil, err
			}
			return &endpoint.RawMsg, nil
		}
	}

	return nil, ErrChannelNotFound
}

func (h *Handler) channelKey(
	logger log.Logger, channel *data.Channel) (data.Base64String, error) {
	offering, err := h.offeringByID(logger, channel.Offering)
	if err != nil {
		return "", err
	}
	keyBytes, err := data.ChannelKey(
		channel.Client, channel.Agent, channel.Block, offering.Hash)
	if err != nil {
		logger.Error(err.Error())
		return "", ErrInternal
	}
	return data.FromBytes(keyBytes), nil
}
