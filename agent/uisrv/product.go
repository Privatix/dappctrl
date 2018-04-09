package uisrv

import (
	"net/http"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

// handleProducts calls appropriate handler by scanning incoming request.
func (s *Server) handleProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		s.handlePostProducts(w, r)
		return
	}
	if r.Method == "PUT" {
		s.handlePutProducts(w, r)
		return
	}
	if r.Method == "GET" {
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
	if err := s.db.Insert(product); err != nil {
		s.logger.Warn("failed to insert product: %v", err)
		s.replyUnexpectedErr(w)
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

// handleGetProducts replies with all products.
func (s *Server) handleGetProducts(w http.ResponseWriter, r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params: nil,
		View:   data.ProductTable,
	})
}
