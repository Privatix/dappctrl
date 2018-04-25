// +build !noethtest

package eth

import (
	"testing"
)

func TestNegativeLogsFetching(t *testing.T) {
	client := getClient()
	for _, topics := range [][]string{
		[]string{"0x0"},
		[]string{"0x0"},
		[]string{"", ""},
	} {
		_, err := client.GetLogs("", topics, "", "")
		if err == nil {
			t.Fatal("error expected, got nil")
		}
	}
}
