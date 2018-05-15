// +build !noethtest

package eth

import (
	"testing"
)

func TestBlockNumberFetching(t *testing.T) {
	response, err := getClient().GetBlockNumber()
	if err != nil {
		t.Fatal(err)
	}

	if response.Result == "" {
		t.Fatal("Unexpected response received")
	}
}
