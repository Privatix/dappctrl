package uisrv

import (
	"net/http"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
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
	product := &data.Product{}
	if !s.parseProductPayload(w, r, product) {
		return
	}
	product.ID = util.NewUUID()
	if !s.insert(w, product) {
		return
	}
	s.replyEntityCreated(w, product.ID)
}

// handlePutProducts updates a product.
func (s *Server) handlePutProducts(w http.ResponseWriter, r *http.Request) {
	product := &data.Product{}
	if !s.parseProductPayload(w, r, product) {
		return
	}
	if err := s.db.Update(product); err != nil {
		s.logger.Warn("failed to update product: %v", err)
		s.replyUnexpectedErr(w)
		return
	}
	s.replyEntityUpdated(w, product.ID)
}

func (s *Server) parseProductPayload(w http.ResponseWriter,
	r *http.Request, product *data.Product) bool {
	if !s.parsePayload(w, r, product) ||
		validate.Struct(product) != nil ||
		product.OfferTplID == nil ||
		product.OfferAccessID == nil ||
		(product.UsageRepType != data.ProductUsageIncremental &&
			product.UsageRepType != data.ProductUsageTotal) {
		s.replyInvalidPayload(w)
		return false
	}
	return true
}

// handleGetProducts replies with all products available to the agent.
func (s *Server) handleGetProducts(w http.ResponseWriter, r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params:       nil,
		View:         data.ProductTable,
		FilteringSQL: `products.is_server`,
	})
}

// handleGetProducts replies with all products available to the client.
func (s *Server) handleGetClientProducts(w http.ResponseWriter,
	r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params:       nil,
		View:         data.ProductTable,
		FilteringSQL: `NOT products.is_server`,
	})
}
