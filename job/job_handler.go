package job

import (
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job/handler"
	"github.com/privatix/dappctrl/job/queue"
)

// HandlersMap returns handlers map needed to construct job queue.
func HandlersMap(handler *handler.Handler) queue.HandlerMap {
	// TODO: add clients
	return queue.HandlerMap{
		// Agent jobs.
		data.JobAgentAfterChannelCreate:             handler.AgentAfterChannelCreate,
		data.JobAgentAfterChannelTopUp:              handler.AgentAfterChannelTopUp,
		data.JobAgentAfterUncooperativeCloseRequest: handler.AgentAfterUncooperativeClose,
		data.JobAgentAfterUncooperativeClose:        handler.AgentAfterUncooperativeClose,
		data.JobAgentPreCooperativeClose:            handler.AgentPreCooperativeClose,
		data.JobAgentAfterCooperativeClose:          handler.AgentAfterCooperativeClose,
		data.JobAgentPreServiceSuspend:              handler.AgentPreServiceSuspend,
		data.JobAgentPreServiceUnsuspend:            handler.AgentPreServiceUnsuspend,
		data.JobAgentPreServiceTerminate:            handler.AgentPreServiceTerminate,
		data.JobAgentPreEndpointMsgCreate:           handler.AgentPreEndpointMsgCreate,
		data.JobAgentPreEndpointMsgSOMCPublish:      handler.AgentPreEndpointMsgSOMCPublish,
		data.JobAgentAfterEndpointMsgSOMCPublish:    handler.AgentAfterEndpointMsgSOMCPublish,
		data.JobAgentPreOfferingMsgBCPublish:        handler.AgentPreOfferingMsgBCPublish,
		data.JobAgentAfterOfferingMsgBCPublish:      handler.AgentAfterOfferingMsgBCPublish,
		data.JobAgentPreOfferingMsgSOMCPublish:      handler.AgentPreOfferingMsgSOMCPublish,
		// Common jobs.
		data.JobPreAccountAddBalanceApprove: handler.PreAccountAddBalanceApprove,
		data.JobPreAccountAddBalance:        handler.PreAccountAddBalance,
		data.JobAfterAccountAddBalance:      handler.AfterAccountAddBalance,
		data.JobPreAccountReturnBalance:     handler.PreAccountReturnBalance,
		data.JobAfterAccountReturnBalance:   handler.AfterAccountReturnBalance,
	}
}
