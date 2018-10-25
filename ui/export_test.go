package ui

import (
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/proc"
)

func (h *Handler) SetMockQueue(q job.Queue) {
	h.queue = q
}

func (h *Handler) SetMockRole(role string) {
	if role == data.RoleAgent {
		h.agent = true
		return
	}
	h.agent = false
}

func (h *Handler) SetProcessor(processor *proc.Processor) {
	h.processor = processor
}
