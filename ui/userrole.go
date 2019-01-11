package ui

// GetUserRole returns user role.
func (h *Handler) GetUserRole(tkn string) (*string, error) {
	logger := h.logger.Add("method", "GetUserRole")

	if !h.token.Check(tkn) {
		logger.Warn("access denied")
		return nil, ErrAccessDenied
	}

	return &h.userRole, nil
}
