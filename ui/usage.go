package ui

// GetOfferingUsage returns total units used for all channels
// with a given offering.
func (h *Handler) GetOfferingUsage(password, offeringID string) (*uint, error) {
	logger := h.logger.Add("method", "GetOfferingUsage",
		"offeringID", offeringID)

	return h.uintFromQuery(logger, password,
		`SELECT SUM(sessions.units_used)
		   FROM channels
		   	JOIN sessions
			ON sessions.channel=channels.id
			   AND channels.offering=$1`,
		offeringID)
}

// GetProductUsage returns total units used in all channel
// of all offerings with given product.
func (h *Handler) GetProductUsage(password, productID string) (*uint, error) {
	logger := h.logger.Add("method", "GetProductUsage",
		"productID", productID)

	return h.uintFromQuery(logger, password,
		`SELECT SUM(sessions.units_used)
		   FROM offerings
		   	JOIN channels
			ON channels.offering=offerings.id
			   AND offerings.product=$1
		   	JOIN sessions
		     	ON sessions.channel=channels.id`, productID)
}
