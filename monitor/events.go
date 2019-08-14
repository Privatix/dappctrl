package monitor

import (
	"github.com/ethereum/go-ethereum/common"
)

const (
	logChannelCreated            = "LogChannelCreated"
	logChannelToppedUp           = "LogChannelToppedUp"
	logChannelCloseRequested     = "LogChannelCloseRequested"
	logOfferingCreated           = "LogOfferingCreated"
	logOfferingDeleted           = "LogOfferingDeleted"
	logOfferingPopedUp           = "LogOfferingPopedUp"
	logCooperativeChannelClose   = "LogCooperativeChannelClose"
	logUnCooperativeChannelClose = "LogUnCooperativeChannelClose"
	approval                     = "Approval"
	transfer                     = "Transfer"
)

func (m *Monitor) eventNameFromHash(hash common.Hash) string {
	switch hash {
	case m.pscABI.Events[logChannelCreated].ID():
		return logChannelCreated
	case m.pscABI.Events[logChannelToppedUp].ID():
		return logChannelToppedUp
	case m.pscABI.Events[logChannelCloseRequested].ID():
		return logChannelCloseRequested
	case m.pscABI.Events[logOfferingCreated].ID():
		return logOfferingCreated
	case m.pscABI.Events[logOfferingDeleted].ID():
		return logOfferingDeleted
	case m.pscABI.Events[logOfferingPopedUp].ID():
		return logOfferingPopedUp
	case m.pscABI.Events[logCooperativeChannelClose].ID():
		return logCooperativeChannelClose
	case m.pscABI.Events[logUnCooperativeChannelClose].ID():
		return logUnCooperativeChannelClose
	case m.ptcABI.Events[approval].ID():
		return approval
	case m.ptcABI.Events[transfer].ID():
		return transfer
	}
	return hash.String()
}
