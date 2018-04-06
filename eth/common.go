package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/privatix/dappctrl/data"
)

// BalanceClosingHash computes balance message hash.
func BalanceClosingHash(clientAddr, pscAddr common.Address, block uint32,
	offeringHash common.Hash, balance *big.Int) []byte {
	blockBytes := data.Uint32ToBytes(block)
	return crypto.Keccak256(
		crypto.Keccak256(
			[]byte("string message_id"),
			[]byte("address sender"),
			[]byte("uint32 block_created"),
			[]byte("bytes32 offering_hash"),
			[]byte("uint192 balance"),
			[]byte("address contrac"),
		),
		crypto.Keccak256(
			[]byte("Receiver closing signature"),
			clientAddr.Bytes(),
			blockBytes[:],
			offeringHash.Bytes(),
			balance.Bytes(),
			pscAddr.Bytes(),
		),
	)
}

// BalanceProofHash implementes hash as in psc contract.
func BalanceProofHash(clientAddr, pscAddr, agentAddr common.Address, block uint32,
	offeringHash common.Hash, balance *big.Int) []byte {
	blockBytes := data.Uint32ToBytes(block)
	return crypto.Keccak256(
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
			agentAddr.Bytes(),
			blockBytes[:],
			offeringHash.Bytes(),
			balance.Bytes(),
			pscAddr.Bytes(),
		),
	)
}
