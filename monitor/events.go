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
	case m.pscABI.Events[logChannelCreated].Id():
		return logChannelCreated
	case m.pscABI.Events[logChannelToppedUp].Id():
		return logChannelToppedUp
	case m.pscABI.Events[logChannelCloseRequested].Id():
		return logChannelCloseRequested
	case m.pscABI.Events[logOfferingCreated].Id():
		return logOfferingCreated
	case m.pscABI.Events[logOfferingDeleted].Id():
		return logOfferingDeleted
	case m.pscABI.Events[logOfferingPopedUp].Id():
		return logOfferingPopedUp
	case m.pscABI.Events[logCooperativeChannelClose].Id():
		return logCooperativeChannelClose
	case m.pscABI.Events[logUnCooperativeChannelClose].Id():
		return logUnCooperativeChannelClose
	case m.ptcABI.Events[approval].Id():
		return approval
	case m.ptcABI.Events[transfer].Id():
		return transfer
	}
	return hash.String()
}
