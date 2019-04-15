package nfts

// Transaction tags for NFT messages
var (
	TxCategory = "nfts"

	TagSender    = "sender" // TODO: use sdk.TagSender
	TagRecipient = "recipient"
	TagOwner     = "owner"
	TagNFTID     = "nft-id"
	TagDenom     = "denom"
	TagCategory  = "category" // TODO: use sdk.TagCategory
)
