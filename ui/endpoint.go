package ui

import (
	"strings"

	"github.com/privatix/dappctrl/data"
)

func (h *Handler) getEndpointsConditions(
	channel, template string) (tail string, args []interface{}) {
	var conditions []string

	if channel != "" {
		conditions = append(conditions, "channel")
		args = append(args, channel)
	}

	if template != "" {
		conditions = append(conditions, "template")
		args = append(args, template)
	}

	items := h.tailElements(conditions)

	if len(items) > 0 {
		tail = "WHERE " + strings.Join(items, " AND ")
	}

	return tail, args
}

// GetEndpoints returns endpoints.
func (h *Handler) GetEndpoints(
	tkn, channel, template string) ([]data.Endpoint, error) {
	logger := h.logger.Add("method", "GetEndpoints",
		"channel", channel, "template", template)

	if !h.token.Check(tkn) {
		return nil, ErrAccessDenied
	}

	tail, args := h.getEndpointsConditions(channel, template)

	result, err := h.selectAllFrom(
		logger, data.EndpointTable, tail, args...)
	if err != nil {
		return nil, err
	}

	endpoints := make([]data.Endpoint, len(result))
	for i, item := range result {
		endpoints[i] = *item.(*data.Endpoint)
	}

	return endpoints, nil
}
