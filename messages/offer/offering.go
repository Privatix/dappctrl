package offer

import (
	"github.com/privatix/dappctrl/data"
)

// OfferingMessage returns new Offering message
func OfferingMessage(agent *data.Account, template *data.Template,
	offering *data.Offering) *Message {
	msg := &Message{
		AgentPubKey:               agent.PublicKey,
		TemplateHash:              template.Hash,
		Country:                   offering.Country,
		ServiceSupply:             offering.Supply,
		UnitName:                  offering.UnitName,
		UnitType:                  offering.UnitType,
		BillingType:               offering.BillingType,
		SetupPrice:                offering.SetupPrice,
		UnitPrice:                 offering.UnitPrice,
		MinUnits:                  offering.MinUnits,
		MaxUnit:                   offering.MaxUnit,
		BillingInterval:           offering.BillingInterval,
		MaxBillingUnitLag:         offering.MaxBillingUnitLag,
		MaxSuspendTime:            offering.MaxSuspendTime,
		MaxInactiveTimeSec:        offering.MaxInactiveTimeSec,
		FreeUnits:                 offering.FreeUnits,
		ServiceSpecificParameters: offering.AdditionalParams,
	}
	return msg
}
