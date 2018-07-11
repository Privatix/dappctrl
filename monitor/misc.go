package monitor

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func mustParseABI(abiJSON string) (abi.ABI, error) {
	return abi.JSON(strings.NewReader(abiJSON))
}
