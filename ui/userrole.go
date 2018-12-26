package ui

// GetUserRole returns user role.
func (h *Handler) GetUserRole(tkn string) (*string, error) {
	if !h.token.Check(tkn) {
		return nil, ErrAccessDenied
	}

	return &h.userRole, nil
}
