package data

// Job creators.
const (
	JobUser           = "user"
	JobBillingChecker = "billing_checker"
	JobBCMonitor      = "bc_monitor"
	JobTask           = "task"
	JobServiceAdapter = "service_adapter"
	JobSessionServer  = "session_server"
)

// Job statuses.
const (
	JobActive   = "active"
	JobDone     = "done"
	JobFailed   = "failed"
	JobCanceled = "canceled"
)

// Job related object types.
const (
	JobOffering = "offering"
	JobChannel  = "channel"
	JobEndpoint = "endpoint"
	JobAccount  = "account"
)

// Job types.
const (
	JobClientPreChannelCreate               = "clientPreChannelCreate"
	JobClientAfterChannelCreate             = "clientAfterChannelCreate"
	JobClientPreChannelTopUp                = "clientPreChannelTopUp"
	JobClientAfterChannelTopUp              = "clientAfterChannelTopUp"
	JobClientPreUncooperativeCloseRequest   = "clientPreUncooperativeCloseRequest"
	JobClientAfterUncooperativeCloseRequest = "clientAfterUncooperativeCloseRequest"
	JobClientPreUncooperativeClose          = "clientPreUncooperativeClose"
	JobClientAfterUncooperativeClose        = "clientAfterUncooperativeClose"
	JobClientAfterCooperativeClose          = "clientAfterCooperativeClose"
	JobClientPreServiceSuspend              = "clientPreServiceSuspend"
	JobClientPreServiceUnsuspend            = "clientPreServiceUnsuspend"
	JobClientPreServiceTerminate            = "clientPreServiceTerminate"
	JobClientEndpointRestore                = "clientEndpointRestore"
	JobClientAfterOfferingMsgBCPublish      = "clientAfterOfferingMsgBCPublish"
	JobClientAfterOfferingPopUp             = "clientAfterOfferingPopUp"
	JobClientAfterOfferingDelete            = "clientAfterOfferingDelete"
	JobClientCompleteServiceTransition      = "completeServiceTransition"
	JobAgentAfterChannelCreate              = "agentAfterChannelCreate"
	JobAgentAfterChannelTopUp               = "agentAfterChannelTopUp"
	JobAgentAfterUncooperativeCloseRequest  = "agentAfterUncooperativeCloseRequest"
	JobAgentAfterUncooperativeClose         = "agentAfterUncooperativeClose"
	JobAgentAfterCooperativeClose           = "agentAfterCooperativeClose"
	JobAgentPreServiceSuspend               = "agentPreServiceSuspend"
	JobAgentPreServiceUnsuspend             = "agentPreServiceUnsuspend"
	JobAgentPreServiceTerminate             = "agentPreServiceTerminate"
	JobAgentPreEndpointMsgCreate            = "agentPreEndpointMsgCreate"
	JobAgentPreOfferingMsgBCPublish         = "agentPreOfferingMsgBCPublish"
	JobAgentAfterOfferingMsgBCPublish       = "agentAfterOfferingMsgBCPublish"
	JobAgentPreOfferingDelete               = "agentPreOfferingDelete"
	JobAgentPreOfferingPopUp                = "agentPreOfferingPopUp"
	JobAgentAfterOfferingPopUp              = "agentAfterOfferingPopUp"
	JobAgentAfterOfferingDelete             = "agentAfterOfferingDelete"
	JobPreAccountAddBalanceApprove          = "preAccountAddBalanceApprove"
	JobPreAccountAddBalance                 = "preAccountAddBalance"
	JobAfterAccountAddBalance               = "afterAccountAddBalance"
	JobPreAccountReturnBalance              = "preAccountReturnBalance"
	JobAfterAccountReturnBalance            = "afterAccountReturnBalance"
	JobAccountUpdateBalances                = "accountUpdateBalances"
	JobDecrementCurrentSupply               = "decrementCurrentSupply"
	JobIncrementCurrentSupply               = "incrementCurrentSupply"
)

// JobEthLog is log data a job derived from.
type JobEthLog struct {
	Block  uint64    `json:"block"`
	Data   []byte    `json:"data"`
	Topics LogTopics `json:"topics"`
	TxHash HexString `json:"transactionHash"`
}

// JobData data set by blockchain monitor for log derived jobs.
type JobData struct {
	EthLog *JobEthLog `json:"ethereumLog"`
}

// JobBalanceData is a data required for transfer jobs.
type JobBalanceData struct {
	GasPrice uint64
	Amount   uint64
}

// JobPublishData is a data required for blockchain publish jobs.
type JobPublishData struct {
	GasPrice uint64
}

// JobTopUpChannelData is a data for top up channel job.
type JobTopUpChannelData struct {
	GasPrice uint64
	Deposit  uint64
}

// JobCreateChannelData is a data required by client to accept an offering.
type JobCreateChannelData struct {
	GasPrice uint64
	Deposit  uint
}

// JobEndpointCreateData is a data for client endpoint create job.
type JobEndpointCreateData struct {
	EndpointSealed []byte
}
