package data

import (
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/reform.v1"
)

// FromBase64ToHex return hex of base 64 encoded.
func FromBase64ToHex(s string) (string, error) {
	b, err := ToBytes(s)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HexToBytes reutrns the bytes represented by the hex of string s.
func HexToBytes(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

// HexFromBytes returns the hex encoding of src.
func HexFromBytes(src []byte) string {
	return hex.EncodeToString(src)
}

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
	if err != nil {
		err = fmt.Errorf("unable to parse ethereum hash: %s", err)
	}
	return common.BytesToHash(hashBytes), err
}

// HexToHash returns the ethereum's hash represented by the hex of string s.
func HexToHash(h string) (common.Hash, error) {
	hashBytes, err := HexToBytes(h)
	if err != nil {
		err = fmt.Errorf("unable to parse ethereum hash: %s", err)
	}
	return common.BytesToHash(hashBytes), err
}

// HexToAddress returns ethereum's address from base 64 encoded string.
func HexToAddress(addr string) (common.Address, error) {
	addrBytes, err := HexToBytes(addr)
	if err != nil {
		err = fmt.Errorf("unable to parse ethereum addr: %s", err)
	}
	return common.BytesToAddress(addrBytes), err
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

// Uint192ToBytes using big endian with leading zeros.
func Uint192ToBytes(x *big.Int) [24]byte {
	var ret [24]byte
	xBytes := x.Bytes()
	for i, v := range xBytes {
		ret[24-len(xBytes)+i] = v
	}
	return ret
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

// IsAgent specifies user role. "true" - agent. "false" - client.
func IsAgent(db *reform.Querier) (bool, error) {
	var setting Setting
	err := db.FindByPrimaryKeyTo(&setting, IsAgentKey)
	if err != nil {
		var err2 error

		if err == sql.ErrNoRows {
			err2 = fmt.Errorf("key %s is not exist"+
				" in Setting table", IsAgentKey)
		} else {
			err2 = fmt.Errorf("failed to get key %s"+
				" from Setting table", IsAgentKey)
		}
		return false, err2
	}
	val, err := strconv.ParseBool(setting.Value)
	if err != nil {
		return false, fmt.Errorf("key %s from Setting table"+
			" has an incorrect format", IsAgentKey)
	}
	return val, nil
}

// FindByPrimaryKeyTo calls db.FindByPrimaryKeyTo() returning more descriptive
// error.
func FindByPrimaryKeyTo(db *reform.Querier,
	rec reform.Record, key interface{}) error {
	if err := db.FindByPrimaryKeyTo(rec, key); err != nil {
		return fmt.Errorf("failed to find %T by primary key: %s",
			rec, err)
	}
	return nil
}

// Insert calls db.Insert() returning more descriptive error.
func Insert(db *reform.Querier, str reform.Struct) error {
	if err := db.Insert(str); err != nil {
		return fmt.Errorf("failed to insert %T: %s", str, err)
	}
	return nil
}

// Save calls db.Save() returning more descriptive error.
func Save(db *reform.Querier, rec reform.Record) error {
	if err := db.Save(rec); err != nil {
		return fmt.Errorf("failed to save %T: %s", rec, err)
	}
	return nil
}

// FindOneTo calls db.FindOneTo() returning more descriptive error.
func FindOneTo(db *reform.Querier,
	str reform.Struct, column string, arg interface{}) error {
	if err := db.FindOneTo(str, column, arg); err != nil {
		return fmt.Errorf("failed to find %T by '%s' column: %s",
			str, column, err)
	}
	return nil
}

// ChannelKey returns the unique channel identifier
// used in a Privatix Service Contract.
func ChannelKey(client, agent string, block uint32,
	offeringHash string) ([]byte, error) {
	clientAddr, err := HexToAddress(client)
	if err != nil {
		return nil, err
	}

	agentAddr, err := HexToAddress(agent)
	if err != nil {
		return nil, err
	}

	hash, err := base64.URLEncoding.DecodeString(
		strings.TrimSpace(offeringHash))
	if err != nil {
		return nil, err
	}

	blockBytes := Uint32ToBytes(block)

	return crypto.Keccak256(clientAddr.Bytes(),
		agentAddr.Bytes(), blockBytes[:],
		common.BytesToHash(hash).Bytes()), nil
}

// MinDeposit calculates minimal deposit required to accept the offering.
func MinDeposit(offering *Offering) uint64 {
	return offering.MinUnits*offering.UnitPrice + offering.SetupPrice
}
