package tags

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Transaction tags for NFT messages
var (
	TxCategory = "nfts"

	Category  = sdk.TagCategory
	Sender    = sdk.TagSender
	Recipient = "recipient"
	Owner     = "owner"
	NFTID     = "nft-id"
	Denom     = "denom"
)
