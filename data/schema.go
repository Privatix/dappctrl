package data

import (
	"time"
)

//go:generate reform

// Account is an ethereum account.
//reform:accounts
type Account struct {
	ID               string     `json:"id" reform:"id,pk"`
	EthAddr          string     `json:"ethAddr" reform:"eth_addr"`
	PublicKey        string     `json:"-" reform:"public_key"`
	PrivateKey       string     `json:"-" reform:"private_key"`
	IsDefault        bool       `json:"isDefault" reform:"is_default"`
	InUse            bool       `json:"inUse" reform:"in_use"`
	Name             string     `json:"name" reform:"name"`
	PTCBalance       uint64     `json:"ptcBalance" reform:"ptc_balance"`
	PSCBalance       uint64     `json:"psc_balance" reform:"psc_balance"`
	EthBalance       B64BigInt  `json:"ethBalance" reform:"eth_balance"`
	LastBalanceCheck *time.Time `json:"lastBalanceCheck" reform:"last_balance_check"`
}

// User is party in distributed trade.
// It can play an agent role, a client role, or both of them.
//reform:users
type User struct {
	ID        string `json:"id" reform:"id,pk"`
	EthAddr   string `json:"ethAddr" reform:"eth_addr"`
	PublicKey string `json:"publicKey" reform:"public_key"`
}

// Templates kinds.
const (
	TemplateOffer  = "offer"
	TemplateAccess = "access"
)

// Template is a user defined structures.
// It can be an offer or access template.
//reform:templates
type Template struct {
	ID   string `json:"id" reform:"id,pk"`
	Hash string `json:"hash" reform:"hash"`
	Raw  []byte `json:"raw" reform:"raw"`
	Kind string `json:"kind" reform:"kind"`
}

// Product usage reporting types.
const (
	ProductUsageIncremental = "incremental"
	ProductUsageTotal       = "total"
)

// Product authentication types.
const (
	ClientIdentByChannelID = "by_channel_id"
)

// Product stores billing and action related settings.
//reform:products
type Product struct {
	ID                     string  `json:"id" reform:"id,pk"`
	Name                   string  `json:"name" reform:"name"`
	OfferTplID             *string `json:"offerTplID" reform:"offer_tpl_id"`
	OfferAccessID          *string `json:"offerAccessID" reform:"offer_access_id"`
	UsageRepType           string  `json:"usageRepType" reform:"usage_rep_type"`
	IsServer               bool    `json:"isServer" reform:"is_server"`
	Salt                   uint64  `json:"-" reform:"salt"`
	Password               string  `json:"-" reform:"password"`
	ClientIdent            string  `json:"clientIdent" reform:"client_ident"`
	Config                 []byte  `json:"config" reform:"config"`
	ServiceEndpointAddress *string `json:"serviceEndpointAddress" reform:"service_endpoint_address"`
}

// Unit used for billing calculation.
const (
	UnitScalar  = "units"
	UnitSeconds = "seconds"
)

// Billing types.
const (
	BillingPrepaid  = "prepaid"
	BillingPostpaid = "postpaid"
)

// Message statuses.
const (
	MsgUnpublished      = "unpublished"           // Saved but not published.
	MsgBChainPublishing = "bchain_publishing"     // To blockchain.
	MsgBChainPublished  = "bchain_published"      // To blockchain.
	MsgChPublished      = "msg_channel_published" // Published in messaging channel.
)

// Offering statuses.
const (
	OfferEmpty    = "empty"
	OfferRegister = "register"
	OfferRemove   = "remove"
)

// Offering is a service offering.
//reform:offerings
type Offering struct {
	ID                 string  `json:"id" reform:"id,pk"`
	IsLocal            bool    `json:"is_local" reform:"is_local"`
	Template           string  `json:"template" reform:"tpl" validate:"required"`    // Offering's.
	Product            string  `json:"product" reform:"product" validate:"required"` // Specific billing and actions.
	Hash               string  `json:"hash" reform:"hash"`                           // Offering's hash.
	Status             string  `json:"status" reform:"status"`
	OfferStatus        string  `json:"offerStatus" reform:"offer_status"`
	BlockNumberUpdated uint64  `json:"blockNumberUpdated" reform:"block_number_updated"`
	Agent              string  `json:"agent" reform:"agent" validate:"required"`
	RawMsg             string  `json:"rawMsg" reform:"raw_msg"`
	ServiceName        string  `json:"serviceName" reform:"service_name" validate:"required"`
	Description        *string `json:"description" reform:"description"`
	Country            string  `json:"country" reform:"country" validate:"required"` // ISO 3166-1 alpha-2.
	Supply             uint16  `json:"supply" reform:"supply" validate:"required"`
	UnitName           string  `json:"unitName" reform:"unit_name" validate:"required"` // Like megabytes, minutes, etc.
	UnitType           string  `json:"unitType" reform:"unit_type" validate:"required"`
	BillingType        string  `json:"billingType" reform:"billing_type" validate:"required"`
	SetupPrice         uint64  `json:"setupPrice" reform:"setup_price"` // Setup fee.
	UnitPrice          uint64  `json:"unitPrice" reform:"unit_price"`
	MinUnits           uint64  `json:"minUnits" reform:"min_units" validate:"required"`
	MaxUnit            *uint64 `json:"maxUnit" reform:"max_unit"`
	BillingInterval    uint    `json:"billingInterval" reform:"billing_interval" validate:"required"` // Every unit number to be paid.
	MaxBillingUnitLag  uint    `json:"maxBillingUnitLag" reform:"max_billing_unit_lag"`               // Max maximum tolerance for payment lag.
	MaxSuspendTime     uint    `json:"maxSuspendTime" reform:"max_suspended_time"`                    // In seconds.
	MaxInactiveTimeSec *uint64 `json:"maxInactiveTimeSec" reform:"max_inactive_time_sec"`
	FreeUnits          uint8   `json:"freeUnits" reform:"free_units"`
	AdditionalParams   []byte  `json:"additionalParams" reform:"additional_params" validate:"required"`
}

// State channel statuses.
const (
	ChannelPending       = "pending"
	ChannelActive        = "active"
	ChannelWaitCoop      = "wait_coop"
	ChannelClosedCoop    = "closed_coop"
	ChannelWaitChallenge = "wait_challenge"
	ChannelInChallenge   = "in_challenge"
	ChannelWaitUncoop    = "wait_uncoop"
	ChannelClosedUncoop  = "closed_uncoop"
)

// Service operational statuses.
const (
	ServicePending    = "pending"
	ServiceActive     = "active"
	ServiceSuspended  = "suspended"
	ServiceTerminated = "terminated"
)

// Channel is a state channel.
//reform:channels
type Channel struct {
	ID                 string     `json:"id" reform:"id,pk"`
	Agent              string     `json:"agent" reform:"agent"`
	Client             string     `json:"client" reform:"client"`
	Offering           string     `json:"offering" reform:"offering"`
	Block              uint32     `json:"block" reform:"block"`                  // When state channel created.
	ChannelStatus      string     `json:"channelStatus" reform:"channel_status"` // Status related to blockchain.
	ServiceStatus      string     `json:"serviceStatus" reform:"service_status"`
	ServiceChangedTime *time.Time `json:"serviceChangedTime" reform:"service_changed_time"`
	TotalDeposit       uint64     `json:"totalDeposit" reform:"total_deposit"`
	Salt               uint64     `json:"-" reform:"salt"`
	Username           *string    `json:"-" reform:"username"`
	Password           string     `json:"-" reform:"password"`
	ReceiptBalance     uint64     `json:"-" reform:"receipt_balance"`   // Last payment.
	ReceiptSignature   *string    `json:"-" reform:"receipt_signature"` // Last payment's signature.
}

// Session is a client session.
//reform:sessions
type Session struct {
	ID              string     `json:"id" reform:"id,pk"`
	Channel         string     `json:"channel" reform:"channel"`
	Started         time.Time  `json:"started" reform:"started"`
	Stopped         *time.Time `json:"stopped" reform:"stopped"`
	UnitsUsed       uint64     `json:"unitsUsed" reform:"units_used"`
	SecondsConsumed uint64     `json:"secondsConsumed" reform:"seconds_consumed"`
	LastUsageTime   time.Time  `json:"lastUsageTime" reform:"last_usage_time"`
	ClientIP        *string    `json:"clientIP" reform:"client_ip"`
	ClientPort      *uint16    `json:"clientPort" reform:"client_port"`
}

// Contract types.
const (
	ContractPTC = "ptc"
	ContractPSC = "psc"
)

// Contract is a smart contract.
//reform:contracts
type Contract struct {
	ID      string `json:"id" reform:"id,pk"`
	Address string `json:"address" reform:"address"` // Ethereum address
	Type    string `json:"type" reform:"type"`
	Version *uint8 `json:"version" reform:"version"`
	Enabled bool   `json:"enabled" reform:"enabled"`
}

// Setting is a user setting.
//reform:settings
type Setting struct {
	Key         string  `json:"key" reform:"key,pk"`
	Value       string  `json:"value" reform:"value"`
	Description *string `json:"description" reform:"description"`
	Name        string  `json:"name" reform:"name"`
}

// Endpoint messages is info about service access.
//reform:endpoints
type Endpoint struct {
	ID                     string  `json:"id" reform:"id,pk"`
	Template               string  `json:"template" reform:"template"`
	Channel                string  `json:"channel" reform:"channel"`
	Hash                   string  `json:"hash" reform:"hash"`
	RawMsg                 string  `reform:"raw_msg"`
	Status                 string  `json:"status" reform:"status"`
	PaymentReceiverAddress *string `json:"paymentReceiverAddress" reform:"payment_receiver_address"`
	ServiceEndpointAddress *string `json:"serviceEndpointAddress" reform:"service_endpoint_address"`
	Username               *string `json:"username" reform:"username"`
	Password               *string `json:"password" reform:"password"`
	AdditionalParams       []byte  `json:"additionalParams" reform:"additional_params"`
}

// EndpointUI contains only certain fields of endpoints table.
//reform:endpoints
type EndpointUI struct {
	ID               string `json:"id" reform:"id,pk"`
	AdditionalParams []byte `json:"additionalParams" reform:"additional_params"`
}

// Job creators.
const (
	JobUser           = "user"
	JobBillingChecker = "billing_checker"
	JobBCMonitor      = "bc_monitor"
	JobTask           = "task"
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
	JobOfferring = "offering"
	JobChannel   = "channel"
	JobEndpoint  = "endpoint"
	JobAccount   = "account"
)

// Transaction statuses.
const (
	TxUnsent = "unsent"
	TxSent   = "sent"
	TxMined  = "mined"
	TxUncle  = "uncle"
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
	JobClientPreServiceTerminate            = "clientPreServiceTerminate"
	JobClientAfterServiceTerminate          = "clientAfterServiceTerminate"
	JobClientPreEndpointMsgSOMCGet          = "clientPreEndpointMsgSOMCGet"
	JobClientAfterOfferingMsgBCPublish      = "clientAfterOfferingMsgBCPublish"
	JobClientPreOfferingMsgSOMCGet          = "clientPreOfferingMsgSOMCGet"
	JobAgentAfterChannelCreate              = "agentAfterChannelCreate"
	JobAgentAfterChannelTopUp               = "agentAfterChannelTopUp"
	JobAgentAfterUncooperativeCloseRequest  = "agentAfterUncooperativeCloseRequest"
	JobAgentAfterUncooperativeClose         = "agentAfterUncooperativeClose"
	JobAgentPreCooperativeClose             = "agentPreCooperativeClose"
	JobAgentAfterCooperativeClose           = "agentAfterCooperativeClose"
	JobAgentPreServiceSuspend               = "agentPreServiceSuspend"
	JobAgentPreServiceUnsuspend             = "agentPreServiceUnsuspend"
	JobAgentPreServiceTerminate             = "agentPreServiceTerminate"
	JobAgentPreEndpointMsgCreate            = "agentPreEndpointMsgCreate"
	JobAgentPreEndpointMsgSOMCPublish       = "agentPreEndpointMsgSOMCPublish"
	JobAgentAfterEndpointMsgSOMCPublish     = "agentAfterEndpointMsgSOMCPublish"
	JobAgentPreOfferingMsgBCPublish         = "agentPreOfferingMsgBCPublish"
	JobAgentAfterOfferingMsgBCPublish       = "agentAfterOfferingMsgBCPublish"
	JobAgentPreOfferingMsgSOMCPublish       = "agentPreOfferingMsgSOMCPublish"
	JobPreAccountAddBalanceApprove          = "preAccountAddBalanceApprove"
	JobPreAccountAddBalance                 = "preAccountAddBalance"
	JobAfterAccountAddBalance               = "afterAccountAddBalance"
	JobPreAccountReturnBalance              = "preAccountReturnBalance"
	JobAfterAccountReturnBalance            = "afterAccountReturnBalance"
	JobAccountAddCheckBalance               = "addCheckBalance"
)

// JobBalanceData is a data required for transfer jobs.
type JobBalanceData struct {
	GasPrice uint64
	Amount   uint
}

// JobPublishData is a data required for blockchain publish jobs.
type JobPublishData struct {
	GasPrice uint64
}

// Job is a task within persistent queue.
//reform:jobs
type Job struct {
	ID          string    `reform:"id,pk"`
	Type        string    `reform:"type"`
	Status      string    `reform:"status"`
	RelatedType string    `reform:"related_type"`
	RelatedID   string    `reform:"related_id"`
	CreatedAt   time.Time `reform:"created_at"`
	NotBefore   time.Time `reform:"not_before"`
	CreatedBy   string    `reform:"created_by"`
	TryCount    uint8     `reform:"try_count"`
	Data        []byte    `reform:"data"`
}

// EthTx is an ethereum transaction
//reform:eth_txs
type EthTx struct {
	ID          string    `reform:"id,pk" json:"id"`
	Hash        string    `reform:"hash" json:"hash"`
	Method      string    `reform:"method" json:"method"`
	Status      string    `reform:"status" json:"status"`
	JobID       *string   `reform:"job" json:"jobID"`
	Issued      time.Time `reform:"issued" json:"issued"`
	AddrFrom    string    `reform:"addr_from" json:"addrFrom"`
	AddrTo      string    `reform:"addr_to" json:"addrTo"`
	Nonce       *string   `reform:"nonce" json:"nonce"`
	GasPrice    uint64    `reform:"gas_price" json:"gasPrice"`
	Gas         uint64    `reform:"gas" json:"gas"`
	TxRaw       []byte    `reform:"tx_raw" json:"txRaw"`
	RelatedType string    `reform:"related_type" json:"relatedType"`
	RelatedID   string    `reform:"related_id" json:"relatedID"`
}

// EthLog is an ethereum log entry.
//reform:eth_logs
type EthLog struct {
	ID          string    `reform:"id,pk"`
	TxHash      string    `reform:"tx_hash"`
	TxStatus    string    `reform:"status"`
	JobID       *string   `reform:"job"`
	BlockNumber uint64    `reform:"block_number"`
	Addr        string    `reform:"addr"`
	Data        string    `reform:"data"`
	Topics      LogTopics `reform:"topics"`
	Failures    uint64    `reform:"failures"`
	Ignore      bool      `reform:"ignore"`
}
