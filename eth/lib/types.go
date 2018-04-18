package lib

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
)

const addressBytesLength = 20

// todo: try to use type Address [20]byte instead of current realisation.
type Address struct {
	bytes [addressBytesLength]byte
}

func NewAddress(hexRepresentation string) (*Address, error) {
	hexSource := hexRepresentation

	if len(hexSource) != addressBytesLength*2 && len(hexSource) != (addressBytesLength*2)+2 { // "0x..."
		return nil, errors.New("address might be decoded from 40 symbols long hex string literal only")
	}

	// In case if address is prefixed with 0x -
	// it should be removed for proper decoding.
	if hexSource[:2] == "0x" {
		hexSource = hexSource[2:]
	}

	decodedAddress, err := hex.DecodeString(hexSource)
	if err != nil {
		return nil, err
	}

	address := &Address{}
	copy(address.bytes[:], decodedAddress[:addressBytesLength])
	return address, nil
}

// Tests: types_test.go/TestAddressCreating
// (this method is used for checking address creation correctness,
// so no separated tests are needed).
func (a *Address) String() string {
	return fmt.Sprintf("%#x", a.bytes)
}

// Tests: types_test.go/TestAddressCreating
// (String() uses a.bytes for hex representation,
// in case if it is malformed - String() tests would fail).
func (a *Address) Bytes() [addressBytesLength]byte {
	return a.bytes
}

// todo: try to use type Uint256s *BigInt instead of current realisation.
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
	if len(hexSource) == 1 {
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

// todo: try to use type Uint192 *BigInt instead of current realisation.
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
	if len(hexSource) == 1 {
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
