package messages

import (
	"encoding/json"
	"reflect"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// Hash computes and returns hash of message.
func Hash(msg interface{}) ([]byte, error) {
	// If msg is a byte slice, do not marshal.
	if reflect.TypeOf(msg) != reflect.SliceOf(reflect.TypeOf(byte(0))) {
		buff, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		return ethcrypto.Keccak256(buff), nil
	}
	return ethcrypto.Keccak256(msg.([]byte)), nil
}
