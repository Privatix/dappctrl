package offer

import (
	"encoding/json"

	"github.com/xeipuuv/gojsonschema"

	"github.com/privatix/dappctrl/data"
)

// OfferingMessage returns new Offering message
func OfferingMessage(agent *data.Account, template *data.Template,
	offering *data.Offering) *Message {
	msg := &Message{
		AgentPubKey:        agent.PublicKey,
		TemplateHash:       template.Hash,
		Country:            offering.Country,
		ServiceSupply:      offering.Supply,
		UnitName:           offering.UnitName,
		UnitType:           offering.UnitType,
		BillingType:        offering.BillingType,
		SetupPrice:         offering.SetupPrice,
		UnitPrice:          offering.UnitPrice,
		MinUnits:           offering.MinUnits,
		MaxUnit:            offering.MaxUnit,
		BillingInterval:    offering.BillingInterval,
		MaxBillingUnitLag:  offering.MaxBillingUnitLag,
		MaxSuspendTime:     offering.MaxSuspendTime,
		MaxInactiveTimeSec: offering.MaxInactiveTimeSec,
		FreeUnits:          offering.FreeUnits,
		Nonce:              offering.ID,
		ServiceSpecificParameters: offering.AdditionalParams,
	}
	return msg
}

// ValidMsg if is true then offering message corresponds
// to an offer template scheme.
func ValidMsg(schema json.RawMessage, msg Message) bool {
	sch := gojsonschema.NewBytesLoader(schema)
	loader := gojsonschema.NewGoLoader(msg)

	result, err := gojsonschema.Validate(sch, loader)
	if err != nil || !result.Valid() || len(result.Errors()) != 0 {
		return false
	}
	return true
}
