package mon

import "github.com/privatix/dappctrl/util/errors"

// Monitor errors.
const (
	// CRC16("github.com/privatix/dappctrl/svc/dappvpn") = 0x0604
	ErrServerOutdated errors.Error = 0x0604<<8 + iota
)

var errMsgs = errors.Messages{
	ErrServerOutdated: "server outdated",
}

func init() {
	errors.InjectMessages(errMsgs)
}
