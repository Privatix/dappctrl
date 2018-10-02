package sesssrv

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/privatix/dappctrl/country"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/util/srv"
)

const (
	externalIP = "externalIP"
	defaultIP  = "127.0.0.1"
)

// ProductArgs is a set of product arguments.
type ProductArgs struct {
	Config map[string]string `json:"config"`
}

func (s *Server) handleProductConfig(
	w http.ResponseWriter, r *http.Request, ctx *srv.Context) {
	logger := s.logger.Add("method", "handleProductConfig",
		"sender", r.RemoteAddr)

	logger.Info("session product config request from " + r.RemoteAddr)

	var args ProductArgs
	if !s.ParseRequest(logger, w, r, &args) {
		return
	}

	if len(args.Config) == 0 {
		s.RespondError(logger, w, ErrInvalidProductConf)
		return
	}

	product, ok := s.findProduct(logger, w, ctx.Username)
	if !ok {
		return
	}

	product.ServiceEndpointAddress = serviceEndpointAddress(args.Config,
		product)

	delete(args.Config, externalIP)

	if product.ServiceEndpointAddress != nil {
		c := findCountry(
			s.countryConf, *product.ServiceEndpointAddress, logger)
		if len(c) != 2 {
			c = country.UndefinedCountry
		}
		product.Country = &c
	}

	prodConf, err := json.Marshal(args.Config)
	if err != nil {
		logger.Error(err.Error())
		s.RespondError(logger, w, srv.ErrInternalServerError)
		return
	}

	product.Config = prodConf

	if ok := s.updateProduct(logger, w, product); !ok {
		return

	}

	s.RespondResult(logger, w, nil)
}

func findCountry(config *country.Config,
	ip string, logger log.Logger) string {
	url := strings.Replace(config.URLTemplate, "{{ip}}", ip, 1)

	logger = logger.Add("url", url, "field", config.Field,
		"timeout", config.Timeout, "ip", ip)

	c, err := country.GetCountry(
		config.Timeout, url, config.Field)
	if err != nil {
		logger.Error(err.Error())
		return country.UndefinedCountry
	}
	return c
}

func serviceEndpointAddress(config map[string]string,
	product *data.Product) *string {
	if product == nil {
		return nil
	}

	if (product.ServiceEndpointAddress != nil &&
		*product.ServiceEndpointAddress != defaultIP) ||
		config == nil {
		return product.ServiceEndpointAddress
	}

	if extIP, ok := config[externalIP]; ok {
		ip := net.ParseIP(extIP)
		if ip != nil {
			return &extIP
		}
		return product.ServiceEndpointAddress
	}

	return product.ServiceEndpointAddress
}
