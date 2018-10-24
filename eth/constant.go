package eth

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Logs hexes.
var (
	// PSC logs.
	ServiceChannelCreated            = common.HexToHash("78ef868cbadb8c73fc1170b8a09d99704c4cdf400c9c94534811b0ae60b8913a")
	ServiceChannelToppedUp           = common.HexToHash("a3b2cd532a9050531ecc674928d7704894707ede1a436bfbee86b96b83f2a5ce")
	ServiceChannelCloseRequested     = common.HexToHash("b40564b1d36572b2942ad7cfc5a5a967f3ef08c82163a910dee760c5b629a32e")
	ServiceOfferingCreated           = common.HexToHash("98fd7525022191e0cb78e4245bba1fd0f2be10e6a7bf8d640151ca9d005ae73b")
	ServiceOfferingDeleted           = common.HexToHash("c3013cd9dd5c33b95a9cc1bc076481c9a6a1970be6d7f1ed33adafad6e57d3d6")
	ServiceOfferingPopedUp           = common.HexToHash("3db9eb24898d4657e5f8f72b8584750dba18dcdc974e8cf0fa9507dc44e2e174")
	ServiceCooperativeChannelClose   = common.HexToHash("b488ea0f49970f556cf18e57588e78dcc1d3fd45c71130aa5099a79e8b06c8e7")
	ServiceUnCooperativeChannelClose = common.HexToHash("7418f9b30b6de272d9d54ee6822f674042c58cea183b76d5d4e7b3c933a158f6")

	// PTC logs.
	TokenApproval = common.HexToHash("8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925")
	TokenTransfer = common.HexToHash("ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
)

// BlockDuration is an average block duration.
const BlockDuration = time.Duration(15) * time.Second
