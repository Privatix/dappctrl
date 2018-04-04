package data

import "time"

//go:generate reform

// User is party in distributed trade.
// It can play an agent role, a client role, or both of them.
//reform:users
type User struct {
	ID         string  `json:"id" reform:"id,pk"`
	PublicKey  string  `json:"publicKey" reform:"public_key"`
	PrivateKey *string `json:"privateKey" reform:"private_key"`
	Default    bool    `json:"default" reform:"is_default"`
	InUse      bool    `json:"inUse" reform:"in_use"`
}

// Templates kinds.
const (
	TemplateOffer  = "offer"
	TemplateAccess = "access"
)

// Template is a user defined structures.
// It can be an offer, auth or access template.
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

// Product stores billing and action related settings.
//reform:products
type Product struct {
	ID            string  `json:"id" reform:"id,pk"`
	Name          string  `json:"name" reform:"name"`
	OfferTplID    *string `json:"offerTplID" reform:"offer_tpl_id"`
	OfferAccessID *string `json:"offerAccessID" reform:"offer_access_id"`
	UsageRepType  string  `json:"usageRepType" reform:"usage_rep_type"`
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

// Offering is a service offering.
//reform:offerings
type Offering struct {
	ID                 string  `json:"id" reform:"id,pk"`
	Template           string  `json:"template" reform:"tpl" validate:"required"`    // Offering's.
	Product            string  `json:"product" reform:"product" validate:"required"` // Specific billing and actions.
	Hash               string  `json:"hash" reform:"hash"`                           // Offering's hash.
	Status             string  `json:"status" reform:"status"`
	Agent              string  `json:"agent" reform:"agent" validate:"required"`
	Signature          string  `json:"signature" reform:"signature"` // Agent's signature.
	ServiceName        string  `json:"serviceName" reform:"service_name" validate:"required"`
	Description        *string `json:"description" reform:"description"`
	Country            string  `json:"country" reform:"country" validate:"required"` // ISO 3166-1 alpha-2.
	Supply             uint    `json:"supply" reform:"supply" validate:"required"`
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
	Nonce              string  `json:"nonce" reform:"nonce"`
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
	Block              uint       `json:"block" reform:"block"`                  // When state channel created.
	ChannelStatus      string     `json:"channelStatus" reform:"channel_status"` // Status related to blockchain.
	ServiceStatus      string     `json:"serviceStatus" reform:"service_status"`
	ServiceChangedTime *time.Time `json:"serviceChangedTime" reform:"service_changed_time"`
	TotalDeposit       uint64     `json:"totalDeposit" reform:"total_deposit"`
	Salt               uint64     `json:"-" reform:"salt"`
	Username           *string    `json:"-" reform:"username"`
	Password           string     `json:"-" reform:"password"`
	ReceiptBalance     uint64     `json:"-" reform:"receipt_balance"`   // Last payment.
	ReceiptSignature   string     `json:"-" reform:"receipt_signature"` // Last payment's signature.
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
	ServerIP        *string    `json:"serverIP" reform:"server_ip"`
	ServerPort      *int16     `json:"serverPort" reform:"server_port"`
	ClientIP        *string    `json:"clientIP" reform:"client_ip"`
	ClientPort      *int16     `json:"clientPort" reform:"client_port"`
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
}

// Endpoint messages is info about service access.
//reform:endpoints
type Endpoint struct {
	ID                     string  `json:"id" reform:"id,pk"`
	Template               string  `json:"template" reform:"tpl"`
	Channel                string  `json:"channel" reform:"channel"`
	Hash                   string  `json:"hash" reform:"hash"`
	Status                 string  `json:"status" reform:"status"`
	Signature              string  `json:"signature" reform:"signature"`
	PaymentReceiverAddress *string `json:"paymentReceiverAddress" reform:"payment_receiver_address"`
	DNS                    *string `json:"dns" reform:"dns"`
	IPAddress              *string `json:"ipAddress" reform:"ip_addr"`
	Username               *string `json:"-" reform:"username"`
	Password               *string `json:"-" reform:"password"`
	AdditionalParams       []byte  `json:"additionalParams" reform:"additional_params"`
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
	JobStatusNew     = "new"
	JobStatusFailed  = "failed"
	JobStatusSkipped = "skipped"
	JobStatusDone    = "done"
)

// Transaction statuses.
const (
	TxUnsent = "unsent"
	TxSent   = "sent"
	TxMined  = "mined"
	TxUncle  = "uncle"
)
