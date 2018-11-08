package somc

import "github.com/privatix/dappctrl/data"

// Offering returns offerings raw msg with given hash.
func (h *Handler) Offering(hash data.HexString) (*data.Base64String, error) {
	logger := h.logger.Add("method", "Offering")

	offering, err := h.offeringByHash(logger, hash)
	if err != nil {
		if err == ErrOfferingNotFound {
			logger.Warn("unexpected request for offering: " + string(hash))
		}
		return nil, err
	}

	return &offering.RawMsg, nil
}
