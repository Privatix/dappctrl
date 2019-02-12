package ui

import (
	"strings"

	"github.com/privatix/dappctrl/data"
)

func (h *Handler) getSessionsConditions(
	channel string) (tail string, args []interface{}) {
	var conditions []string

	if channel != "" {
		conditions = append(conditions, "channel")
		args = append(args, channel)
	}

	items := h.tailElements(conditions)
	if len(items) > 0 {
		tail = "WHERE " + strings.Join(items, " AND ")
	}
	return tail, args
}

// GetSessions returns sessions.
func (h *Handler) GetSessions(tkn, channel string) ([]data.Session, error) {
	logger := h.logger.Add("method", "GetSessions", "channel", channel)

	if !h.token.Check(tkn) {
		logger.Warn("access denied")
		return nil, ErrAccessDenied
	}

	tail, args := h.getSessionsConditions(channel)

	result, err := h.selectAllFrom(
		logger, data.SessionTable, tail, args...)
	if err != nil {
		return nil, err
	}

	sessions := make([]data.Session, len(result))
	for i, item := range result {
		sessions[i] = *item.(*data.Session)
	}

	return sessions, nil
}
