package uisrv

import (
	"fmt"
	"net/http"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

// handleProducts calls appropriate handler by scanning incoming request.
func (s *Server) handleProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.handlePostProducts(w, r)
		return
	}
	if r.Method == http.MethodPut {
		s.handlePutProducts(w, r)
		return
	}
	if r.Method == http.MethodGet {
		s.handleGetProducts(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handlePostProducts creates new product.
func (s *Server) handlePostProducts(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.Add("method", "handlePostProducts")

	product := &data.Product{}
	if !s.parseProductPayload(logger, w, r, product) {
		return
	}
	product.ID = util.NewUUID()
	if !s.insert(logger, w, product) {
		return
	}
	s.replyEntityCreated(logger, w, product.ID)
}

// handlePutProducts updates a product.
func (s *Server) handlePutProducts(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.Add("method", "handlePutProducts")

	product := &data.Product{}
	if !s.parseProductPayload(logger, w, r, product) {
		return
	}

	// TODO(maxim) make it on front-end
	oldProduct := new(data.Product)
	if err := s.db.FindByPrimaryKeyTo(oldProduct, product.ID); err != nil {
		logger.Warn(fmt.Sprintf("failed to find product: %v", err))
		s.replyUnexpectedErr(logger, w)
		return
	}

	if product.Salt == 0 {
		product.Salt = oldProduct.Salt
	}

	if product.Password == "" {
		product.Password = oldProduct.Password
	}

	if err := s.db.Update(product); err != nil {
		logger.Warn(fmt.Sprintf("failed to update product: %v", err))
		s.replyUnexpectedErr(logger, w)
		return
	}
	s.replyEntityUpdated(logger, w, product.ID)
}

func (s *Server) parseProductPayload(logger log.Logger, w http.ResponseWriter,
	r *http.Request, product *data.Product) bool {
	if !s.parsePayload(logger, w, r, product) ||
		validate.Struct(product) != nil ||
		product.OfferTplID == nil ||
		product.OfferAccessID == nil ||
		(product.UsageRepType != data.ProductUsageIncremental &&
			product.UsageRepType != data.ProductUsageTotal) {
		s.replyInvalidRequest(logger, w)
		return false
	}
	return true
}

// handleGetProducts replies with all products available to the agent.
func (s *Server) handleGetProducts(w http.ResponseWriter, r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params: nil,
		View:   data.ProductTable,
		FilteringSQL: filteringSQL{
			SQL:      `products.is_server`,
			JoinWith: "ADD",
		},
	})
}

// handleGetProducts replies with all products available to the client.
func (s *Server) handleGetClientProducts(w http.ResponseWriter,
	r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params: nil,
		View:   data.ProductTable,
		FilteringSQL: filteringSQL{
			SQL:      `NOT products.is_server`,
			JoinWith: "ADD",
		},
	})
}
