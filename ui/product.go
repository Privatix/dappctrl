package ui

import (
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

// CreateProduct creates new product.
func (h *Handler) CreateProduct(tkn string,
	product data.Product) (*string, error) {
	logger := h.logger.Add("method", "CreateProduct", "product", product)

	if !h.token.Check(tkn) {
		logger.Warn("access denied")
		return nil, ErrAccessDenied
	}

	if product.ServiceEndpointAddress != nil &&
		!isValidSEAddress(*product.ServiceEndpointAddress) {
		return nil, ErrBadServiceEndpointAddress
	}

	product.ID = util.NewUUID()
	if err := insert(logger, h.db.Querier, &product); err != nil {
		return nil, err
	}

	return &product.ID, nil
}

// UpdateProduct updates a product.
func (h *Handler) UpdateProduct(tkn string, product data.Product) error {
	logger := h.logger.Add("method", "UpdateProduct", "product", product)

	if !h.token.Check(tkn) {
		logger.Warn("access denied")
		return ErrAccessDenied
	}

	oldProduct := &data.Product{}
	if err := h.findByPrimaryKey(logger,
		ErrProductNotFound, oldProduct, product.ID); err != nil {
		return err
	}

	if product.Salt == 0 {
		product.Salt = oldProduct.Salt
	}

	if product.Password == "" {
		product.Password = oldProduct.Password
	}

	if product.ServiceEndpointAddress != nil &&
		!isValidSEAddress(*product.ServiceEndpointAddress) {
		return ErrBadServiceEndpointAddress
	}

	if err := update(logger, h.db.Querier, &product); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return nil
}

func isValidSEAddress(address string) bool {
	if util.IsIPv4(address) || util.IsHostname(address) {
		return true
	}
	return false
}

// GetProducts returns all products available to the agent.
func (h *Handler) GetProducts(tkn string) ([]data.Product, error) {
	logger := h.logger.Add("method", "GetProducts")

	if !h.token.Check(tkn) {
		logger.Warn("access denied")
		return nil, ErrAccessDenied
	}

	result, err := h.db.SelectAllFrom(data.ProductTable,
		"WHERE products.is_server")
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	products := make([]data.Product, len(result))
	for i, item := range result {
		products[i] = *item.(*data.Product)
	}

	return products, nil
}
