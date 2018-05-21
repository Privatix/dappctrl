package data

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"golang.org/x/crypto/sha3"
)

// ToBytes returns the bytes represented by the base64 string s.
func ToBytes(s string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(strings.TrimSpace(s))
}

// FromBytes returns the base64 encoding of src.
func FromBytes(src []byte) string {
	return base64.URLEncoding.EncodeToString(src)
}

// ToHash returns the ethereum's hash represented by the base64 string s.
func ToHash(h string) (common.Hash, error) {
	hashBytes, err := ToBytes(h)
	ret := common.BytesToHash(hashBytes)
	return ret, err
}

// ToAddress returns ethereum's address from base 64 encoded string.
func ToAddress(addr string) (common.Address, error) {
	addrBytes, err := ToBytes(addr)
	ret := common.BytesToAddress(addrBytes)
	return ret, err
}

// BytesToUint32 using big endian.
func BytesToUint32(b []byte) (uint32, error) {
	if len(b) != 4 {
		return 0, fmt.Errorf("wrong len")
	}
	return binary.BigEndian.Uint32(b), nil
}

// Uint32ToBytes using big endian.
func Uint32ToBytes(x uint32) [4]byte {
	var xBytes [4]byte
	binary.BigEndian.PutUint32(xBytes[:], x)
	return xBytes
}

// ValidatePassword checks if a given password, hash and salt are matching.
func ValidatePassword(password, hash string, salt uint64) bool {
	h := sha3.Sum256([]byte(password + fmt.Sprint(salt)))
	return FromBytes(h[:]) == hash
}
