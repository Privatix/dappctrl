package ui

// GetUserRole returns user role.
func (h *Handler) GetUserRole() (*string, error) {
	return &h.userRole, nil
}
