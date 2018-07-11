package eth

import "time"

// Logs digests.
const (
	// PSC logs.
	EthDigestChannelCreated      = "a6153987181667023837aee39c3f1a702a16e5e146323ef10fb96844a526143c"
	EthDigestChannelToppedUp     = "a3b2cd532a9050531ecc674928d7704894707ede1a436bfbee86b96b83f2a5ce"
	EthChannelCloseRequested     = "b40564b1d36572b2942ad7cfc5a5a967f3ef08c82163a910dee760c5b629a32e"
	EthOfferingCreated           = "32c1913dfde418197923027c2f2260f19903a2e86a93ed83c4689ac91a96bafd"
	EthOfferingDeleted           = "c3013cd9dd5c33b95a9cc1bc076481c9a6a1970be6d7f1ed33adafad6e57d3d6"
	EthOfferingEndpoint          = "450e7ab61f9e1c40dd7c79edcba274a7a96f025fab1733b3fa1087a1b5d1db7d"
	EthOfferingPoppedUp          = "c37352067a3ca1eafcf2dc5ba537fc473509c4e4aaca729cb1dab7053ec1ffbf"
	EthCooperativeChannelClose   = "b488ea0f49970f556cf18e57588e78dcc1d3fd45c71130aa5099a79e8b06c8e7"
	EthUncooperativeChannelClose = "7418f9b30b6de272d9d54ee6822f674042c58cea183b76d5d4e7b3c933a158f6"

	// PTC logs.
	EthTokenApproval = "8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"
	EthTokenTransfer = "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
)

// BlockDuration is an average block duration.
const BlockDuration = time.Duration(15) * time.Second
