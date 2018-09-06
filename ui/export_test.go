package ui

import "github.com/privatix/dappctrl/job"

func (h *Handler) SetMockQueue(q job.Queue) {
	h.queue = q
}
