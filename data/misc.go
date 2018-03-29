package data

import (
	"encoding/base64"

	"github.com/ethereum/go-ethereum/common/number"
	"github.com/privatix/dappctrl/util"
)

// ToBytes returns the bytes represented by the base64 string s.
func ToBytes(s string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(s)
}

// FromBytes returns the base64 encoding of src.
func FromBytes(src []byte) string {
	return base64.URLEncoding.EncodeToString(src)
}

// PrivateKeyBytes returns private key's binary representation.
func (s *User) PrivateKeyBytes() ([]byte, error) {
	return ToBytes(*s.PrivateKey)
}

// PublicKeyBytes returns private key's binary representation.
func (s *User) PublicKeyBytes() ([]byte, error) {
	return ToBytes(s.PublicKey)
}

// TotalDepositNum returns total deposit as eth's number.Number.
func (ch *Channel) TotalDepositNum() (*number.Number, error) {
	return util.Base64ToEthNum(ch.TotalDeposit)
}

// ReceiptBalanceNum returns total deposit as eth's number.Number.
func (ch *Channel) ReceiptBalanceNum() (*number.Number, error) {
	return util.Base64ToEthNum(ch.ReceiptBalance)
}
