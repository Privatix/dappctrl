package ui

import (
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

// CreateProduct creates new product.
func (h *Handler) CreateProduct(password string,
	product data.Product) (*string, error) {
	logger := h.logger.Add("method", "CreateProduct", "product", product)

	if err := h.checkPassword(logger, password); err != nil {
		return nil, err
	}

	product.ID = util.NewUUID()
	if err := insert(logger, h.db.Querier, &product); err != nil {
		return nil, err
	}

	return &product.ID, nil
}

// UpdateProduct updates a product.
func (h *Handler) UpdateProduct(password string, product data.Product) error {
	logger := h.logger.Add("method", "UpdateProduct", "product", product)

	if err := h.checkPassword(logger, password); err != nil {
		return err
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

	if err := update(logger, h.db.Querier, &product); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return nil
}

// GetProducts returns all products available to the agent.
func (h *Handler) GetProducts(password string) ([]data.Product, error) {
	logger := h.logger.Add("method", "GetProducts")

	if err := h.checkPassword(logger, password); err != nil {
		return nil, err
	}

	result, err := h.db.SelectAllFrom(data.ProductTable,
		"WHERE products.is_server")
	if err != nil {
		h.logger.Error(err.Error())
		return nil, ErrInternal
	}

	products := make([]data.Product, len(result))
	for i, item := range result {
		products[i] = *item.(*data.Product)
	}

	return products, nil
}
