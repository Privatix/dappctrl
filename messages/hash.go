package messages

import (
	"encoding/json"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// Hash computes and returns hash of message.
func Hash(msg interface{}) ([]byte, error) {
	buff, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return ethcrypto.Keccak256(buff), nil
}
