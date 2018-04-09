package data

import (
	"encoding/base64"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

// ToBytes returns the bytes represented by the base64 string s.
func ToBytes(s string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(strings.TrimSpace(s))
}

// FromBytes returns the base64 encoding of src.
func FromBytes(src []byte) string {
	return base64.URLEncoding.EncodeToString(src)
}

// DecryptPrivateKey is a stub for key decryption.
func DecryptPrivateKey(b []byte) []byte {
	// TODO: implement this.
	return b
}

// Sign signs a data.
func (a *Account) Sign(data []byte) ([]byte, error) {
	prvBytes, err := ToBytes(a.PrivateKey)
	prvBytes = DecryptPrivateKey(prvBytes)
	if err != nil {
		return nil, err
	}
	prv, err := crypto.ToECDSA(prvBytes)
	if err != nil {
		return nil, err
	}
	return crypto.Sign(data, prv)
}

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
		[]byte(offering.Nonce),
		[]byte(offering.Product),
		[]byte(offering.ServiceName),
		big.NewInt(int64(offering.SetupPrice)).Bytes(),
		big.NewInt(int64(offering.Supply)).Bytes(),
		[]byte(offering.Template),
		[]byte(offering.UnitName),
		big.NewInt(int64(offering.UnitPrice)).Bytes(),
	)
}
