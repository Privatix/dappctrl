package somc

import "github.com/privatix/dappctrl/util/errors"

// SOMC errors.
const (
	// CRC16("github.com/privatix/dappctrl/somc") = 0x7CF5
	ErrInternal errors.Error = 0x7CF5<<8 + iota
)

func init() {
	errors.InjectMessages(map[errors.Error]string{
		ErrInternal: "SOMC error",
	})
}
