package proc

import (
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/proc/worker"
)

// HandlersMap returns handlers map needed to construct job queue.
func HandlersMap(worker *worker.Worker) job.HandlerMap {
	// TODO: add clients
	return job.HandlerMap{
		// Agent jobs.
		data.JobAgentAfterChannelCreate:             worker.AgentAfterChannelCreate,
		data.JobAgentAfterChannelTopUp:              worker.AgentAfterChannelTopUp,
		data.JobAgentAfterUncooperativeCloseRequest: worker.AgentAfterUncooperativeClose,
		data.JobAgentAfterUncooperativeClose:        worker.AgentAfterUncooperativeClose,
		data.JobAgentPreCooperativeClose:            worker.AgentPreCooperativeClose,
		data.JobAgentAfterCooperativeClose:          worker.AgentAfterCooperativeClose,
		data.JobAgentPreServiceSuspend:              worker.AgentPreServiceSuspend,
		data.JobAgentPreServiceUnsuspend:            worker.AgentPreServiceUnsuspend,
		data.JobAgentPreServiceTerminate:            worker.AgentPreServiceTerminate,
		data.JobAgentPreEndpointMsgCreate:           worker.AgentPreEndpointMsgCreate,
		data.JobAgentPreEndpointMsgSOMCPublish:      worker.AgentPreEndpointMsgSOMCPublish,
		data.JobAgentAfterEndpointMsgSOMCPublish:    worker.AgentAfterEndpointMsgSOMCPublish,
		data.JobAgentPreOfferingMsgBCPublish:        worker.AgentPreOfferingMsgBCPublish,
		data.JobAgentAfterOfferingMsgBCPublish:      worker.AgentAfterOfferingMsgBCPublish,
		data.JobAgentPreOfferingMsgSOMCPublish:      worker.AgentPreOfferingMsgSOMCPublish,
		// Common jobs.
		data.JobPreAccountAddBalanceApprove: worker.PreAccountAddBalanceApprove,
		data.JobPreAccountAddBalance:        worker.PreAccountAddBalance,
		data.JobAfterAccountAddBalance:      worker.AfterAccountAddBalance,
		data.JobPreAccountReturnBalance:     worker.PreAccountReturnBalance,
		data.JobAfterAccountReturnBalance:   worker.AfterAccountReturnBalance,
		data.JobAccountAddCheckBalance:      worker.AccountAddCheckBalance,
	}
}
