package messages

import "github.com/privatix/dappctrl/util/errors"

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/client/svcrun") = 0x7FDE
	ErrWrongSignature errors.Error = 0x7FDE<<8 + iota
)

func init() {
	errors.InjectMessages(map[errors.Error]string{
		ErrWrongSignature: "wrong signature",
	})
}
