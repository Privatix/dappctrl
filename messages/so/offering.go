package so

import (
	"github.com/privatix/dappctrl/data"
)

// NewOfferingMessage returns new Offering message
func NewOfferingMessage(agent *data.Account, template *data.Template,
	offering *data.Offering) *OfferingMessage {
	msg := &OfferingMessage{
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
