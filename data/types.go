package data

import (
	"fmt"
	"math/big"
)

// B64BigInt is a base64 of big.Int that implements json.Marshaler.
type B64BigInt string

// MarshalJSON marshals itself.
func (i B64BigInt) MarshalJSON() ([]byte, error) {
	buf, err := ToBytes(string(i))
	if err != nil {
		return nil, fmt.Errorf("could not decode base64: %v", err)
	}
	v := big.NewInt(0)
	v.SetBytes(buf)
	return []byte(v.String()), nil
}
