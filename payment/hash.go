package payment

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
)

// Computes hash of a payload.
func hash(pld *payload) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], pld.Balance)
	data := crypto.Keccak256(
		crypto.Keccak256(
			[]byte("string message_id"),
			[]byte("address receiver"),
			[]byte("uint32 block_created"),
			[]byte("bytes32 offering_hash"),
			[]byte("uint192 balance"),
			[]byte("address contract"),
		),
		crypto.Keccak256(
			[]byte("Privatix: sender balance proof signature"),
			[]byte(pld.AgentAddress),
			big.NewInt(int64(pld.OpenBlockNumber)).Bytes(),
			[]byte(pld.OfferingHash),
			buf[:],
			[]byte(pld.ContractAddress),
		),
	)
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	return crypto.Keccak256([]byte(msg))
}
