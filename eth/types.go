package eth

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// NewAddress returns ethereum's common.Address from given hex string.
func NewAddress(addrHex string) (addr common.Address, err error) {
	if len(addrHex) != common.AddressLength*2 && len(addrHex) != (common.AddressLength*2)+2 { // "0x..."
		err = errors.New("address might be decoded from 40 symbols long hex string literal only")
		return
	}
	return common.HexToAddress(addrHex), nil
}

// TODO: try to use type Uint256s *BigInt instead of current implementation.

type Uint256 struct {
	number *big.Int
}

func NewUint256(hexRepresentation string) (*Uint256, error) {
	hexSource := hexRepresentation

	// Hex representation might be shorter, than 64 symbols,
	// but must not be longer than 64 symbols.
	const hexRepresentationLength = 256 / 8 * 2
	if len(hexSource) == 0 || len(hexSource) > 2+hexRepresentationLength {
		return nil, errors.New("uint256 might be decoded from strings like 0x{64 symbols}")
	}

	if len(hexSource) == 2+hexRepresentationLength && hexSource[:2] != "0x" {
		return nil, errors.New("uint256 might be decoded from strings like 0x{64 symbols}")
	}

	// In case if value is prefixed with 0x -
	// it should be removed for proper decoding.
	if len(hexSource) >= 2 && hexSource[:2] == "0x" {
		hexSource = hexSource[2:]
	}

	// In some cases, hex representation might omit leading zeroes,
	// for example 0x0 should be 0x00 in the correct representation.
	// This correction is needed, because otherwise hex.DecodeString() would fail with an error.
	if len(hexSource)%2 == 1 {
		hexSource = "0" + hexSource
	}

	b, err := hex.DecodeString(hexSource)
	if err != nil {
		return nil, err
	}

	i := big.NewInt(0).SetBytes(b)
	return &Uint256{number: i}, nil
}

func (i *Uint256) String() string {
	return fmt.Sprintf("%#x", i.number)
}

func (i *Uint256) ToBigInt() *big.Int {
	return i.number
}

// TODO: try to use type Uint192 *BigInt instead of current realisation.

type Uint192 struct {
	number *big.Int
}

func NewUint192(hexRepresentation string) (*Uint192, error) {
	hexSource := hexRepresentation

	// Hex representation might be shorter, than 48 symbols,
	// but must not be longer than 42 symbols.
	const hexRepresentationLength = 192 / 8 * 2
	if len(hexSource) == 0 || len(hexSource) > 2+hexRepresentationLength {
		return nil, errors.New("uint256 might be decoded from strings like 0x{48 symbols}")
	}

	if len(hexSource) == 2+hexRepresentationLength && hexSource[:2] != "0x" {
		return nil, errors.New("uint256 might be decoded from strings like 0x{48 symbols}")
	}

	// In case if value is prefixed with 0x -
	// it should be removed for proper decoding.
	if len(hexSource) >= 2 && hexSource[:2] == "0x" {
		hexSource = hexSource[2:]
	}

	// In some cases, hex representation might omit leading zeroes,
	// for example 0x0 should be 0x00 in the correct representation.
	// This correction is needed, because otherwise hex.DecodeString() would fail with an error.
	if len(hexSource)%2 == 1 {
		hexSource = "0" + hexSource
	}

	b, err := hex.DecodeString(hexSource)
	if err != nil {
		print(err)
		return nil, err
	}

	i := big.NewInt(0).SetBytes(b)
	return &Uint192{number: i}, nil
}

func (i *Uint192) String() string {
	return fmt.Sprintf("%#x", i.number)
}

func (i *Uint192) ToBigInt() *big.Int {
	return i.number
}
