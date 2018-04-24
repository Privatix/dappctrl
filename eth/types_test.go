// +build !noethtest

package eth

import "testing"

func TestUint256Creating(t *testing.T) {
	checkValidHex := func(hex string, comparableRepresentation string) {
		b, err := NewUint256(hex)
		if err != nil {
			t.Fatal("uint256 variable should be created well, but error catched: ", err)
		}

		if b.String() != hex && b.String() != comparableRepresentation {
			t.Fatal("uint256 variable was encoded to different string representation")
		}
	}

	checkInvalidHex := func(hex string) {
		_, err := NewUint256(hex)
		if err == nil {
			t.Fatal("Error must be returned")
		}
	}

	{
		// Test purpose:
		// To check uint256 decoding, using valid hex representation.
		checkValidHex("0xaaaaffffaaaaffffaaaaffffaaaaffffaaaaffffaaaaffffaaaaffffaaaaffff", "")
	}

	{
		// Test purpose:
		// To check uint256 decoding, using valid zeroed hex representation.
		checkValidHex("0x0000000000000000000000000000000000000000000000000000000000000000", "0x0")
	}

	{
		// Test purpose:
		// To check uint256 decoding, using valid max available hex representation.
		checkValidHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", "")
	}

	{
		// Test purpose:
		// To check uint192 decoding, using valid hex representations,
		// that contains only 4 bits (one symbol only after 0x).
		checkValidHex("0x0", "")
		checkValidHex("0x1", "")
		checkValidHex("0x2", "")
		checkValidHex("0x9", "")
	}

	{
		// Test purpose:
		// To check uint256 decoding, using broken hex representation (longer, than needed).
		// Error must be returned.
		checkInvalidHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffaaaa")
	}
}

func TestUint192Creating(t *testing.T) {
	checkValidHex := func(hex string, comparableRepresentation string) {
		b, err := NewUint192(hex)
		if err != nil {
			t.Fatal("uint192 variable should be created well, but error catched: ", err)
		}

		if b.String() != hex && b.String() != comparableRepresentation {
			t.Fatal("uint192 variable was encoded to different string representation")
		}
	}

	checkInvalidHex := func(hex string) {
		_, err := NewUint192(hex)
		if err == nil {
			t.Fatal("Error must be returned")
		}
	}

	{
		// Test purpose:
		// To check uint192 decoding, using valid hex representation.
		checkValidHex("0xaaaaffffaaaaffffaaaaffffaaaaffffaaaaffffaaaaffff", "")
	}

	{
		// Test purpose:
		// To check uint192 decoding, using valid zeroed hex representation,
		// that would be converted to the 0x0
		checkValidHex("0x000000000000000000000000000000000000000000000000", "0x0")
	}

	{
		// Test purpose:
		// To check uint192 decoding, using valid hex representations,
		// that contains only 4 bits (one symbol only after 0x).
		checkValidHex("0x0", "")
		checkValidHex("0x1", "")
		checkValidHex("0x2", "")
		checkValidHex("0x9", "")
	}

	{
		// Test purpose:
		// To check bytes32 decoding, using valid max possible hex representation.
		checkValidHex("0xffffffffffffffffffffffffffffffffffffffffffffffff", "")
	}

	{
		// Test purpose:
		// To check bytes32 decoding, using broken hex representation (longer, than needed).
		// Error must be returned.
		checkInvalidHex("0xffffffffffffffffffffffffffffffffffffffffffffffaaaa")
	}
}
