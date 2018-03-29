package data

import "time"

//go:generate reform

// User is party in distributed trade.
// It can play an agent role, a client role, or both of them.
//reform:users
type User struct {
	ID         string  `reform:"id,pk"`
	PublicKey  string  `reform:"public_key"`
	PrivateKey *string `reform:"private_key"`
	Default    *bool   `reform:"is_default"`
	InUse      *bool   `reform:"in_use"`
}

// Templates kinds.
const (
	TemplateOffer  = "offer"
	TemplateAuth   = "auth"
	TemplateAccess = "access"
)

// Template is a user defined structures.
// It can be an offer, auth or access template.
//reform:templates
type Template struct {
	ID   string `reform:"id,pk"`
	Hash string `reform:"hash"`
	Raw  []byte `reform:"raw"`
	Kind string `reform:"kind"`
}

// Product uage reporting types.
const (
	ProductUsageIncremental = "incremental"
	ProductUsageTotal       = "total"
)

// Product stores billing and action related settings.
//reform:products
type Product struct {
	ID            string  `reform:"id,pk"`
	Name          string  `reform:"name"`
	OfferTplID    *string `reform:"offer_tpl_id"`
	OfferAccessID *string `reform:"offer_access_id"`
	UsageRepType  string  `reform:"usage_rep_type"`
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
	ID                 string  `reform:"id,pk"`
	Template           string  `reform:"tpl"`     // Offering's.
	Product            string  `reform:"product"` // Specific billing and actions.
	Hash               string  `reform:"hash"`    // Offering's hash.
	Status             string  `reform:"status"`
	Agent              string  `reform:"agent"`
	Signature          string  `reform:"signature"` // Agent's signature.
	TemplateVersion    uint    `reform:"tpl_version"`
	ServiceName        string  `reform:"service_name"`
	Description        *string `reform:"description"`
	Country            string  `reform:"country"` // ISO 3166-1 alpha-2.
	Supply             uint    `reform:"supply"`
	UnitName           string  `reform:"unit_name"` // Like megabytes, minutes, etc.
	UnitType           string  `reform:"unit_type"`
	BillingType        string  `reform:"billing_type"`
	SetupPrice         *string `reform:"setup_price"` // Setup fee.
	UnitPrice          string  `reform:"unit_price"`
	MinUnits           uint64  `reform:"min_units"`
	MaxUnit            *uint64 `reform:"max_unit"`
	BillingInterval    uint    `reform:"billing_interval"`     // Every unit number to be paid.
	MaxBillingUnitLag  uint    `reform:"max_billing_unit_lag"` // Max maximum tolerance for payment lag.
	MaxSuspendTime     uint    `reform:"max_suspended_time"`   // In seconds.
	MaxInactiveTimeSec *uint64 `reform:"max_inactive_time_sec"`
	FreeUints          uint8   `reform:"free_units"`
	Nonce              string  `reform:"nonce"`
	AdditionalParams   []byte  `reform:"additional_params"`
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
	ID                 string     `reform:"id,pk"`
	Agent              string     `reform:"agent"`
	Client             string     `reform:"client"`
	Offering           string     `reform:"offering"`
	Block              uint       `reform:"block"`          // When state channel created.
	ChannelStatus      string     `reform:"channel_status"` // Status related to blockchain.
	ServiceStatus      string     `reform:"service_status"`
	ServiceChangedTime *time.Time `reform:"service_changed_time"`
	TotalDeposit       string     `reform:"total_deposit"`
	Salt               uint64     `reform:"salt"`
	Username           *string    `reform:"username"`
	Password           string     `reform:"password"`
	ReceiptBalance     string     `reform:"receipt_balance"`   // Last payment.
	ReceiptSignature   string     `reform:"receipt_signature"` // Last payment's signature.
}

// Session is a client session.
//reform:sessions
type Session struct {
	ID              string     `reform:"id,pk"`
	Channel         string     `reform:"channel"`
	Started         time.Time  `reform:"started"`
	Stopped         *time.Time `reform:"stopped"`
	UnitsUsed       uint64     `reform:"units_used"`
	SecondsConsumed uint64     `reform:"seconds_consumed"`
	LastUsageTime   time.Time  `reform:"last_usage_time"`
	ServerIP        *string    `reform:"server_ip"`
	ServerPort      *int16     `reform:"server_port"`
	ClientIP        *string    `reform:"client_ip"`
	ClientPort      *int16     `reform:"client_port"`
}

// Contract types.
const (
	ContractPTC = "ptc"
	ContractPSC = "psc"
)

// Contract is a smart contract.
//reform:contracts
type Contract struct {
	ID      string `reform:"id,pk"`
	Address string `reform:"address"` // Ethereum address
	Type    string `reform:"type"`
	Version *uint8 `reform:"version"`
	Enabled bool   `reform:"enabled"`
}

// Endpoint messages is info about service access.
type Endpoint struct {
	ID                     string  `reform:"id,pk"`
	Template               string  `reform:"tpl"`
	Channel                string  `reform:"channel"`
	Hash                   string  `reform:"hash"`
	Status                 string  `reform:"status"`
	Signature              string  `reform:"signature"`
	TemplateVersion        string  `reform:"tplVersion"`
	PaymentReceiverAddress *string `reform:"payment_receiver_address"`
	DNS                    *string `reform:"dns"`
	IPAddress              *string `reform:"ip_addr"`
	Username               *string `reform:"username"`
	Password               *string `reform:"password"`
	AdditionalParams       []byte  `reform:"additional_params"`
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
