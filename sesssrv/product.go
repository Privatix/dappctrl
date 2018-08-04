package sesssrv

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/privatix/dappctrl/data"
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
	var args ProductArgs
	if !s.ParseRequest(w, r, &args) {
		return
	}

	if len(args.Config) == 0 {
		s.RespondError(w, ErrInvalidProductConf)
		return
	}

	product, ok := s.findProduct(w, ctx.Username)
	if !ok {
		return
	}

	product.ServiceEndpointAddress = serviceEndpointAddress(args.Config,
		product)

	delete(args.Config, externalIP)

	prodConf, err := json.Marshal(args.Config)
	if err != nil {
		s.logger.Error(err.Error())
		s.RespondError(w, srv.ErrInternalServerError)
		return
	}

	product.Config = prodConf

	if ok := s.updateProduct(w, product); !ok {
		return

	}

	s.RespondResult(w, nil)
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
