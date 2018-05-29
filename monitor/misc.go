package monitor

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func mustParseABI(abiJSON string) (abi.ABI, error) {
	return abi.JSON(strings.NewReader(abiJSON))
}

func (m *Monitor) errWrapper(ctx context.Context, err error) {
	select {
	case <-ctx.Done():
	default:
		m.errors <- err
	}
}
