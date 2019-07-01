package types

// NFT module event types
var (
	EventTypeTransfer        = "transfer_nft"
	EventTypeEditNFTMetadata = "edit_nft_metadata"
	EventTypeMintNFT         = "mint_nft"
	EventTypeBurnNFT         = "burn_nft"

	AttributeValueCategory = ModuleName

	AttributeKeySender         = "sender"
	AttributeKeyRecipient      = "recipient"
	AttributeKeyOwner          = "owner"
	AttributeKeyNFTID          = "nft-id"
	AttributeKeyNFTName        = "nft-name"
	AttributeKeyNFTDescription = "nft-description"
	AttributeKeyNFTImage       = "nft-image"
	/* #nosec */
	AttributeKeyNFTTokenURI = "nft-tokenURI"
	AttributeKeyDenom       = "denom"
)
