package handlers

import (
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/proc/worker"
)

// HandlersMap returns handlers map needed to construct job queue.
func HandlersMap(worker *worker.Worker) job.HandlerMap {
	return job.HandlerMap{
		// Agent jobs.
		data.JobAgentAfterChannelCreate:             worker.AgentAfterChannelCreate,
		data.JobAgentAfterChannelTopUp:              worker.AgentAfterChannelTopUp,
		data.JobAgentAfterUncooperativeCloseRequest: worker.AgentAfterUncooperativeCloseRequest,
		data.JobAgentAfterUncooperativeClose:        worker.AgentAfterUncooperativeClose,
		data.JobAgentAfterCooperativeClose:          worker.AgentAfterCooperativeClose,
		data.JobAgentPreServiceSuspend:              worker.AgentPreServiceSuspend,
		data.JobAgentPreServiceUnsuspend:            worker.AgentPreServiceUnsuspend,
		data.JobAgentPreServiceTerminate:            worker.AgentPreServiceTerminate,
		data.JobAgentPreEndpointMsgCreate:           worker.AgentPreEndpointMsgCreate,
		data.JobAgentPreOfferingMsgBCPublish:        worker.AgentPreOfferingMsgBCPublish,
		data.JobAgentAfterOfferingMsgBCPublish:      worker.AgentAfterOfferingMsgBCPublish,
		data.JobAgentPreOfferingPopUp:               worker.AgentPreOfferingPopUp,
		data.JobAgentAfterOfferingPopUp:             worker.AgentAfterOfferingPopUp,
		data.JobAgentPreOfferingDelete:              worker.AgentPreOfferingDelete,
		data.JobAgentAfterOfferingDelete:            worker.AgentAfterOfferingDelete,

		// Client jobs.
		data.JobClientAfterOfferingDelete:            worker.ClientAfterOfferingDelete,
		data.JobClientAfterOfferingPopUp:             worker.ClientAfterOfferingPopUp,
		data.JobClientPreChannelCreate:               worker.ClientPreChannelCreate,
		data.JobClientAfterChannelCreate:             worker.ClientAfterChannelCreate,
		data.JobClientEndpointGet:                    worker.ClientEndpointGet,
		data.JobClientAfterUncooperativeClose:        worker.ClientAfterUncooperativeClose,
		data.JobClientAfterCooperativeClose:          worker.ClientAfterCooperativeClose,
		data.JobClientPreUncooperativeClose:          worker.ClientPreUncooperativeClose,
		data.JobClientPreChannelTopUp:                worker.ClientPreChannelTopUp,
		data.JobClientAfterChannelTopUp:              worker.ClientAfterChannelTopUp,
		data.JobClientPreUncooperativeCloseRequest:   worker.ClientPreUncooperativeCloseRequest,
		data.JobClientAfterUncooperativeCloseRequest: worker.ClientAfterUncooperativeCloseRequest,
		data.JobClientPreServiceTerminate:            worker.ClientPreServiceTerminate,
		data.JobClientPreServiceSuspend:              worker.ClientPreServiceSuspend,
		data.JobClientPreServiceUnsuspend:            worker.ClientPreServiceUnsuspend,
		data.JobClientAfterOfferingMsgBCPublish:      worker.ClientAfterOfferingMsgBCPublish,
		data.JobCompleteServiceTransition:            worker.CompleteServiceTransition,
		data.JobClientRecordClosing:                  worker.ClientRecordClosing,

		// Common jobs.
		data.JobPreAccountAddBalanceApprove:   worker.PreAccountAddBalanceApprove,
		data.JobAfterAccountAddBalanceApprove: worker.AfterAccountAddBalanceApprove,
		data.JobPreAccountAddBalance:          worker.PreAccountAddBalance,
		data.JobAfterAccountAddBalance:        worker.AfterAccountAddBalance,
		data.JobPreAccountReturnBalance:       worker.PreAccountReturnBalance,
		data.JobAfterAccountReturnBalance:     worker.AfterAccountReturnBalance,
		data.JobAccountUpdateBalances:         worker.AccountUpdateBalances,
		data.JobDecrementCurrentSupply:        worker.DecrementCurrentSupply,
		data.JobIncrementCurrentSupply:        worker.IncrementCurrentSupply,
	}
}
