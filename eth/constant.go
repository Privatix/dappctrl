package eth

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Logs hexes.
var (
	// PSC logs.
	ServiceChannelCreated            = common.HexToHash("7699dfbb1101aec58b1cfb4a4f5375947c99cb4c645112290d1cb77fc286edc1")
	ServiceChannelToppedUp           = common.HexToHash("d3020c549112ceb2d0f806cd3366f47df57bc519c46133b84ce2cdad970c22a3")
	ServiceChannelCloseRequested     = common.HexToHash("e4007f6ff086417a4031cdcff1a975882379b19e1ed3547292e131f3c525bcab")
	ServiceOfferingCreated           = common.HexToHash("a8c40ba917b58ddcfe866c5b52d417e5e425c459c3b7333bf3b1164e32ddb939")
	ServiceOfferingDeleted           = common.HexToHash("c3013cd9dd5c33b95a9cc1bc076481c9a6a1970be6d7f1ed33adafad6e57d3d6")
	ServiceOfferingPopedUp           = common.HexToHash("7c6ee8c3412a9ecfd989aa18d379f84f73b718366934885e21e9a399b719c53a")
	ServiceCooperativeChannelClose   = common.HexToHash("4a06175bd19aba21163e3c08e7ac80151fad270655624167c5ee9e41b48a58e0")
	ServiceUnCooperativeChannelClose = common.HexToHash("d633584c5931ade1a274b7ce309d985207494d074d1afd2f2da5275bb645e3dc")

	// PTC logs.
	TokenApproval = common.HexToHash("8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925")
	TokenTransfer = common.HexToHash("ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
)

// BlockDuration is an average block duration.
const BlockDuration = time.Duration(15) * time.Second
