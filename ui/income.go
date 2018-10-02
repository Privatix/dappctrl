package ui

// GetOfferingIncome returns total receipt balance from all channels of
// offering with given id.
func (h *Handler) GetOfferingIncome(
	password, offeringID string) (*uint, error) {
	logger := h.logger.Add("method", "GetOfferingIncome",
		"offeringID", offeringID)

	return h.uintFromQuery(logger, password,
		`SELECT SUM(channels.receipt_balance)
		     FROM channels
		   WHERE channels.offering=$1`, offeringID)
}

// GetProductIncome returns total receipt balance from all channels of all
// offerings with given product id.
func (h *Handler) GetProductIncome(
	password, productID string) (*uint, error) {
	logger := h.logger.Add("method", "GetProductIncome",
		"productID", productID)

	return h.uintFromQuery(logger, password,
		`SELECT SUM(channels.receipt_balance)
		     FROM offerings
			  JOIN channels
			  ON channels.offering=offerings.id
		     	     AND offerings.product=$1`, productID)
}
