package data

import (
	"encoding/base64"
	"fmt"
	"strings"

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

// ValidatePassword checks if a given password, hash and salt are matching.
func ValidatePassword(password, hash string, salt uint64) bool {
	h := sha3.Sum256([]byte(password + fmt.Sprint(salt)))
	return base64.URLEncoding.EncodeToString(h[:]) == hash
}
