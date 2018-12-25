package sess

import (
	"encoding/json"
	"net"
	"strings"

	"github.com/privatix/dappctrl/country"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
)

// GetEndpoint returns an endpoint for a given client key.
func (h *Handler) GetEndpoint(
	product, productPassword, clientKey string) (*data.Endpoint, error) {
	logger := h.logger.Add("method", "GetEndpoint",
		"product", product, "clientKey", clientKey)

	logger.Info("channel endpoint request")

	prod, err := h.checkProductPassword(logger, product, productPassword)
	if err != nil {
		return nil, err
	}

	ch, err := h.findClientChannel(logger, prod, clientKey, true)
	if err != nil {
		return nil, err
	}

	var ept data.Endpoint
	if err := h.db.FindOneTo(&ept, "channel", ch.ID); err != nil {
		logger.Error(err.Error())
		return nil, ErrEndpointNotFound
	}

	return &ept, nil
}

// ProductExternalIP is a configuration key for external IP address.
const ProductExternalIP = "externalIP"

func serviceEndpointAddress(
	prod *data.Product, conf map[string]string) *string {
	if prod.ServiceEndpointAddress != nil &&
		*prod.ServiceEndpointAddress != "127.0.0.1" &&
		*prod.ServiceEndpointAddress != "localhost" {
		return prod.ServiceEndpointAddress
	}

	if ipStr, ok := conf[ProductExternalIP]; ok {
		if ip := net.ParseIP(ipStr); ip != nil {
			return &ipStr
		}
	}

	return prod.ServiceEndpointAddress
}

func (h *Handler) findCountry(logger log.Logger, ip string) string {
	url := strings.Replace(h.countryConf.URLTemplate, "{{ip}}", ip, 1)

	logger = logger.Add("url", url, "field", h.countryConf.Field,
		"timeout", h.countryConf.Timeout, "ip", ip)

	country2, err := country.GetCountry(
		h.countryConf.Timeout, url, h.countryConf.Field)
	if err != nil {
		logger.Error(err.Error())
		return country.UndefinedCountry
	}

	return country2
}

// SetProductConfig sets product configuration.
func (h *Handler) SetProductConfig(
	product, productPassword string, config map[string]string) error {
	logger := h.logger.Add("method", "SetProductConfig",
		"product", product, "config", config)

	logger.Info("product config update")

	prod, err := h.checkProductPassword(logger, product, productPassword)
	if err != nil {
		return err
	}

	if len(config) == 0 {
		logger.Warn(ErrBadProductConfig.Error())
		return ErrBadProductConfig
	}

	prod.ServiceEndpointAddress = serviceEndpointAddress(prod, config)
	delete(config, ProductExternalIP)

	if prod.ServiceEndpointAddress != nil {
		country2 := h.findCountry(logger, *prod.ServiceEndpointAddress)
		if len(country2) != 2 {
			country2 = country.UndefinedCountry
		}
		prod.Country = &country2
	}

	if prod.Config, err = json.Marshal(config); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	if err := h.db.Update(prod); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return nil
}
