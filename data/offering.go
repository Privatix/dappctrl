package data

import (
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
)

// OfferingHash returns hash for given Offering.
func OfferingHash(offering *Offering) []byte {
	var maxInactiveTime, maxUnit uint64
	if offering.MaxInactiveTimeSec != nil {
		maxInactiveTime = *offering.MaxInactiveTimeSec
	}
	if offering.MaxUnit != nil {
		maxUnit = *offering.MaxUnit
	}
	return crypto.Keccak256(
		[]byte(offering.ID),
		offering.AdditionalParams,
		[]byte(offering.Agent),
		big.NewInt(int64(offering.BillingInterval)).Bytes(),
		[]byte(offering.BillingType),
		[]byte(offering.Country),
		big.NewInt(int64(offering.FreeUnits)).Bytes(),
		big.NewInt(int64(offering.MaxBillingUnitLag)).Bytes(),
		big.NewInt(int64(maxInactiveTime)).Bytes(),
		big.NewInt(int64(offering.MaxSuspendTime)).Bytes(),
		big.NewInt(int64(maxUnit)).Bytes(),
		big.NewInt(int64(offering.MinUnits)).Bytes(),
		[]byte(offering.Product),
		[]byte(offering.ServiceName),
		big.NewInt(int64(offering.SetupPrice)).Bytes(),
		big.NewInt(int64(offering.Supply)).Bytes(),
		[]byte(offering.Template),
		[]byte(offering.UnitName),
		big.NewInt(int64(offering.UnitPrice)).Bytes(),
	)
}
