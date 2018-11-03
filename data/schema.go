package data

import (
	"encoding/json"
	"time"

	"github.com/privatix/dappctrl/util/log"
)

//go:generate reform

// Account is an ethereum account.
//reform:accounts
type Account struct {
	ID               string       `json:"id" reform:"id,pk"`
	EthAddr          HexString    `json:"ethAddr" reform:"eth_addr"`
	PublicKey        Base64String `json:"-" reform:"public_key"`
	PrivateKey       Base64String `json:"-" reform:"private_key"`
	IsDefault        bool         `json:"isDefault" reform:"is_default"`
	InUse            bool         `json:"inUse" reform:"in_use"`
	Name             string       `json:"name" reform:"name"`
	PTCBalance       uint64       `json:"ptcBalance" reform:"ptc_balance"`
	PSCBalance       uint64       `json:"pscBalance" reform:"psc_balance"`
	EthBalance       Base64BigInt `json:"ethBalance" reform:"eth_balance"`
	LastBalanceCheck *time.Time   `json:"lastBalanceCheck" reform:"last_balance_check"`
}

// User is party in distributed trade.
// It can play an agent role, a client role, or both of them.
//reform:users
type User struct {
	ID        string       `json:"id" reform:"id,pk"`
	EthAddr   HexString    `json:"ethAddr" reform:"eth_addr"`
	PublicKey Base64String `json:"publicKey" reform:"public_key"`
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
	ID   string          `json:"id" reform:"id,pk"`
	Hash HexString       `json:"hash" reform:"hash"`
	Raw  json.RawMessage `json:"raw" reform:"raw"`
	Kind string          `json:"kind" reform:"kind"`
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
	ID                     string          `json:"id" reform:"id,pk"`
	Name                   string          `json:"name" reform:"name"`
	OfferTplID             *string         `json:"offerTplID" reform:"offer_tpl_id"`
	OfferAccessID          *string         `json:"offerAccessID" reform:"offer_access_id"`
	UsageRepType           string          `json:"usageRepType" reform:"usage_rep_type"`
	IsServer               bool            `json:"isServer" reform:"is_server"`
	Salt                   uint64          `json:"-" reform:"salt"`
	Password               Base64String    `json:"-" reform:"password"`
	ClientIdent            string          `json:"clientIdent" reform:"client_ident"`
	Config                 json.RawMessage `json:"config" reform:"config"`
	ServiceEndpointAddress *string         `json:"serviceEndpointAddress" reform:"service_endpoint_address"`
	Country                *string         `json:"country" reform:"country"`
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
	OfferEmpty       = "empty"
	OfferPoppingUp   = "popping_up"
	OfferPoppedUp    = "popped_up"
	OfferRegistering = "registering"
	OfferRegistered  = "registered"
	OfferRemoving    = "removing"
	OfferRemoved     = "removed"
)

// Offering is a service offering.
//reform:offerings
type Offering struct {
	ID                 string          `json:"id" reform:"id,pk"`
	IsLocal            bool            `json:"isLocal" reform:"is_local"`
	Template           string          `json:"template" reform:"tpl" validate:"required"`    // Offering's.
	Product            string          `json:"product" reform:"product" validate:"required"` // Specific billing and actions.
	Hash               HexString       `json:"hash" reform:"hash"`                           // Offering's hash.
	Status             string          `json:"status" reform:"status"`
	OfferStatus        string          `json:"offerStatus" reform:"offer_status"`
	BlockNumberUpdated uint64          `json:"blockNumberUpdated" reform:"block_number_updated"`
	Agent              HexString       `json:"agent" reform:"agent" validate:"required"`
	RawMsg             Base64String    `json:"rawMsg" reform:"raw_msg"`
	ServiceName        string          `json:"serviceName" reform:"service_name" validate:"required"`
	Description        *string         `json:"description" reform:"description"`
	Country            string          `json:"country" reform:"country" validate:"required"` // ISO 3166-1 alpha-2.
	Supply             uint16          `json:"supply" reform:"supply" validate:"required"`
	CurrentSupply      uint16          `json:"currentSupply" reform:"current_supply"`
	UnitName           string          `json:"unitName" reform:"unit_name" validate:"required"` // Like megabytes, minutes, etc.
	UnitType           string          `json:"unitType" reform:"unit_type" validate:"required"`
	BillingType        string          `json:"billingType" reform:"billing_type" validate:"required"`
	SetupPrice         uint64          `json:"setupPrice" reform:"setup_price"` // Setup fee.
	UnitPrice          uint64          `json:"unitPrice" reform:"unit_price"`
	MinUnits           uint64          `json:"minUnits" reform:"min_units" validate:"required"`
	MaxUnit            *uint64         `json:"maxUnit" reform:"max_unit"`
	BillingInterval    uint            `json:"billingInterval" reform:"billing_interval" validate:"required"` // Every unit number to be paid.
	MaxBillingUnitLag  uint            `json:"maxBillingUnitLag" reform:"max_billing_unit_lag"`               // Max maximum tolerance for payment lag.
	MaxSuspendTime     uint            `json:"maxSuspendTime" reform:"max_suspended_time"`                    // In seconds.
	MaxInactiveTimeSec *uint64         `json:"maxInactiveTimeSec" reform:"max_inactive_time_sec"`
	FreeUnits          uint8           `json:"freeUnits" reform:"free_units"`
	AdditionalParams   json.RawMessage `json:"additionalParams" reform:"additional_params" validate:"required"`
	AutoPopUp          *bool           `json:"autoPopUp" reform:"auto_pop_up"`
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
	ID                 string        `json:"id" reform:"id,pk"`
	Agent              HexString     `json:"agent" reform:"agent"`
	Client             HexString     `json:"client" reform:"client"`
	Offering           string        `json:"offering" reform:"offering"`
	Block              uint32        `json:"block" reform:"block"`                  // When state channel created.
	ChannelStatus      string        `json:"channelStatus" reform:"channel_status"` // Status related to blockchain.
	ServiceStatus      string        `json:"serviceStatus" reform:"service_status"`
	ServiceChangedTime *time.Time    `json:"serviceChangedTime" reform:"service_changed_time"`
	PreparedAt         time.Time     `json:"preparedAt" reform:"prepared_at"`
	TotalDeposit       uint64        `json:"totalDeposit" reform:"total_deposit"`
	Salt               uint64        `json:"-" reform:"salt"`
	Username           *string       `json:"-" reform:"username"`
	Password           Base64String  `json:"-" reform:"password"`
	ReceiptBalance     uint64        `json:"receiptBalance" reform:"receipt_balance"` // Last payment.
	ReceiptSignature   *Base64String `json:"-" reform:"receipt_signature"`            // Last payment's signature.
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
	ID      string    `json:"id" reform:"id,pk"`
	Address HexString `json:"address" reform:"address"` // Ethereum address
	Type    string    `json:"type" reform:"type"`
	Version *uint8    `json:"version" reform:"version"`
	Enabled bool      `json:"enabled" reform:"enabled"`
}

// Permissions for settings.
const (
	AccessDenied = iota
	ReadOnly
	ReadWrite
)

// Setting is a user setting.
//reform:settings
type Setting struct {
	Key         string  `json:"key" reform:"key,pk"`
	Value       string  `json:"value" reform:"value"`
	Permissions int     `json:"permissions" reform:"permissions"`
	Description *string `json:"description" reform:"description"`
	Name        string  `json:"name" reform:"name"`
}

// Country statuses.
const (
	CountryStatusUnknown = "unknown"
	CountryStatusValid   = "valid"
	CountryStatusInvalid = "invalid"
)

// Endpoint messages is info about service access.
//reform:endpoints
type Endpoint struct {
	ID                     string       `json:"id" reform:"id,pk"`
	Template               string       `json:"template" reform:"template"`
	Channel                string       `json:"channel" reform:"channel"`
	Hash                   HexString    `json:"hash" reform:"hash"`
	RawMsg                 Base64String `reform:"raw_msg"`
	Status                 string       `json:"status" reform:"status"`
	PaymentReceiverAddress *string      `json:"paymentReceiverAddress" reform:"payment_receiver_address"`
	ServiceEndpointAddress *string      `json:"serviceEndpointAddress" reform:"service_endpoint_address"`
	Username               *string      `json:"username" reform:"username"`
	Password               *string      `json:"password" reform:"password"`
	AdditionalParams       []byte       `json:"additionalParams" reform:"additional_params"`
	CountryStatus          *string      `json:"countryStatus" reform:"country_status"`
}

// EndpointUI contains only certain fields of endpoints table.
//reform:endpoints
type EndpointUI struct {
	ID                     string  `json:"id" reform:"id,pk"`
	PaymentReceiverAddress *string `json:"paymentReceiverAddress" reform:"payment_receiver_address"`
	ServiceEndpointAddress *string `json:"serviceEndpointAddress" reform:"service_endpoint_address"`
	CountryStatus          *string `json:"countryStatus" reform:"country_status"`
}

// Job creators.
const (
	JobUser           = "user"
	JobBillingChecker = "billing_checker"
	JobBCMonitor      = "bc_monitor"
	JobTask           = "task"
	JobServiceAdapter = "service_adapter"
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
	JobClientPreServiceSuspend              = "clientPreServiceSuspend"
	JobClientPreServiceUnsuspend            = "clientPreServiceUnsuspend"
	JobClientPreServiceTerminate            = "clientPreServiceTerminate"
	JobClientPreEndpointMsgSOMCGet          = "clientPreEndpointMsgSOMCGet"
	JobClientAfterEndpointMsgSOMCGet        = "clientAfterEndpointMsgSOMCGet"
	JobClientAfterOfferingMsgBCPublish      = "clientAfterOfferingMsgBCPublish"
	JobClientAfterOfferingPopUp             = "clientAfterOfferingPopUp"
	JobClientAfterOfferingDelete            = "clientAfterOfferingDelete"
	JobAgentAfterChannelCreate              = "agentAfterChannelCreate"
	JobAgentAfterChannelTopUp               = "agentAfterChannelTopUp"
	JobAgentAfterUncooperativeCloseRequest  = "agentAfterUncooperativeCloseRequest"
	JobAgentAfterUncooperativeClose         = "agentAfterUncooperativeClose"
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

type JobEthLog struct {
	Block  uint64    `json:"block"`
	Data   []byte    `json:"data"`
	Topics LogTopics `json:"topics"`
	TxHash HexString `json:"transactionHash"`
}

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

// JobCreateChannelData is a data required by client to accept an offering.
type JobCreateChannelData struct {
	GasPrice uint64
	Deposit  uint
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
	Hash        HexString `reform:"hash" json:"hash"`
	Method      string    `reform:"method" json:"method"`
	Status      string    `reform:"status" json:"status"`
	JobID       *string   `reform:"job" json:"jobID"`
	Issued      time.Time `reform:"issued" json:"issued"`
	AddrFrom    HexString `reform:"addr_from" json:"addrFrom"`
	AddrTo      HexString `reform:"addr_to" json:"addrTo"`
	Nonce       *string   `reform:"nonce" json:"nonce"`
	GasPrice    uint64    `reform:"gas_price" json:"gasPrice"`
	Gas         uint64    `reform:"gas" json:"gas"`
	TxRaw       []byte    `reform:"tx_raw" json:"txRaw"`
	RelatedType string    `reform:"related_type" json:"relatedType"`
	RelatedID   string    `reform:"related_id" json:"relatedID"`
}

// LogEvent is a log event.
//reform:log_events
type LogEvent struct {
	Time    time.Time       `json:"time" reform:"time"`
	Level   log.Level       `json:"level" reform:"level"`
	Message string          `json:"message" reform:"message"`
	Context json.RawMessage `json:"context" reform:"context"`
	Stack   *string         `json:"stack" reform:"stack"`
}
