package offer

import "github.com/privatix/dappctrl/data"

// Message is a message being published to SOMC.
type Message struct {
	AgentPubKey               data.Base64String `json:"agentPublicKey"`
	TemplateHash              data.HexString    `json:"templateHash"`
	Country                   string            `json:"country"`
	ServiceSupply             uint16            `json:"serviceSupply"`
	IPType                    string            `json:"ipType"`
	UnitName                  string            `json:"unitName"`
	UnitType                  string            `json:"unitType"`
	BillingType               string            `json:"billingType"`
	SetupPrice                uint64            `json:"setupPrice"`
	UnitPrice                 uint64            `json:"unitPrice"`
	MinUnits                  uint64            `json:"minUnits"`
	MaxUnit                   *uint64           `json:"maxUnit"`
	BillingInterval           uint              `json:"billingInterval"`
	MaxBillingUnitLag         uint              `json:"maxBillingUnitLag"`
	MaxSuspendTime            uint              `json:"maxSuspendTime"`
	MaxInactiveTimeSec        uint64            `json:"maxInactiveTimeSec"`
	FreeUnits                 uint8             `json:"freeUnits"`
	Nonce                     string            `json:"nonce"`
	ServiceSpecificParameters []byte            `json:"serviceSpecificParameters"`
}
