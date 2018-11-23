package ui

// GetUserRole returns user role.
func (h *Handler) GetUserRole(password string) (*string, error) {
	logger := h.logger.Add("method", "GetUserRole")

	if err := h.checkPassword(logger, password); err != nil {
		return nil, err
	}

	return &h.userRole, nil
}
