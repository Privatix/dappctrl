package data

import (
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/reform.v1"
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

// HashPassword computes encoded hash of the password.
func HashPassword(password, salt string) (string, error) {
	salted := []byte(password + salt)
	passwordHash, err := bcrypt.GenerateFromPassword(salted, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return FromBytes(passwordHash), nil
}

// ValidatePassword checks if a given password, hash and salt are matching.
func ValidatePassword(hash, password, salt string) error {
	salted := []byte(fmt.Sprint(password, salt))
	hashB, err := ToBytes(hash)
	if err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword(hashB, salted)
}

// GetUint64Setting finds the key value in table Setting.
// Checks that the value in the format of uint64
func GetUint64Setting(db *reform.DB, key string) (uint64, error) {
	var setting Setting
	err := db.FindByPrimaryKeyTo(&setting, key)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("key %s is not exist"+
				" in Setting table", key)
		}
		return 0, err
	}

	value, err := strconv.ParseUint(setting.Value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s setting: %v",
			key, err)
	}

	return value, nil

}
