package somc

// Offering returns offerings raw msg with given hash.
func (h *Handler) Offering(hash string) (*string, error) {
	logger := h.logger.Add("method", "Offering")

	offering, err := h.offeringByHash(logger, hash)
	if err != nil {
		if err == ErrOfferingNotFound {
			logger.Warn("unexpected request for offering: " + hash)
		}
		return nil, err
	}

	return &offering.RawMsg, nil
}
