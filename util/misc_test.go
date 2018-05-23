package util

import (
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestMain(m *testing.M) {
	var conf struct{}
	ReadTestConfig(&conf)
	os.Exit(m.Run())
}

func TestGenChannelID(t *testing.T) {
	client := common.HexToAddress("0xa")
	agent := common.HexToAddress("0xb")
	var block uint32 = 0xc
	offering := common.HexToHash("0xd")

	// Got the expected value by calling getKey() from truffle console like this:
	// psc.getKey('0xa', '0xb', 0xc, '\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x0d')
	expected := common.HexToHash("0x2c93cebfce906a0f7aab3b6f6537ae1a8c3621c9f1bc749017da38621b2ea3a0")
	actual := GenChannelID(client, agent, block, offering)
	if actual != expected {
		t.Fatalf("wrong channel id generated: got %s, expected %s", actual.Hex(), expected.Hex())
	}
}
