package data

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Base64String is a base64-encoded binary data.
type Base64String string

// HexString is a hex-encoded binary data.
type HexString string

// Base64BigInt is a base64 of big.Int that implements json.Marshaler.
type Base64BigInt string

// MarshalJSON marshals itself.
func (n Base64BigInt) MarshalJSON() ([]byte, error) {
	buf, err := ToBytes(Base64String(n))
	if err != nil {
		return nil, fmt.Errorf("could not decode base64: %v", err)
	}
	v := big.NewInt(0)
	v.SetBytes(buf)
	return []byte(v.String()), nil
}

// LogTopics is a database/sql compatible type for ethereum log topics.
type LogTopics []common.Hash

// Value serializes the log topics.
func (t LogTopics) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// Scan deserializes the log topics.
func (t *LogTopics) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return fmt.Errorf(
			"type assertion .([]byte) failed, actual type is %T",
			src,
		)
	}

	return json.Unmarshal(source, &t)
}
