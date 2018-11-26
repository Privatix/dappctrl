package ui

import (
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/proc"
)

func (h *Handler) SetMockQueue(q job.Queue) {
	h.queue = q
}

func (h *Handler) SetMockRole(role string) {
	h.userRole = role
}

func (h *Handler) SetProcessor(processor *proc.Processor) {
	h.processor = processor
}
