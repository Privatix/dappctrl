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
	balancePutU192 := data.Uint192ToBytes(balance)
	return crypto.Keccak256(
		[]byte("\x19Ethereum Signed Message:\n32"),
		crypto.Keccak256(
			[]byte("Privatix: receiver closing signature"),
			clientAddr.Bytes(),
			blockBytes[:],
			offeringHash.Bytes(),
			balancePutU192[:],
			pscAddr.Bytes(),
		),
	)
}

// BalanceProofHash implementes hash as in psc contract.
func BalanceProofHash(pscAddr, agentAddr common.Address, block uint32,
	offeringHash common.Hash, balance *big.Int) []byte {
	blockBytes := data.Uint32ToBytes(block)
	balancePutU192 := data.Uint192ToBytes(balance)
	return crypto.Keccak256(
		[]byte("\x19Ethereum Signed Message:\n32"),
		crypto.Keccak256(
			[]byte("Privatix: sender balance proof signature"),
			agentAddr.Bytes(),
			blockBytes[:],
			offeringHash.Bytes(),
			balancePutU192[:],
			pscAddr.Bytes(),
		),
	)
}
