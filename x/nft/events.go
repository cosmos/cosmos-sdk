package nft

// NFT module event types
var (
	EventTypeIssueCollection = "issue_collection"
	EventTypeTransfer        = "transfer_nft"
	EventTypeEditNFT         = "edit_nft"
	EventTypeMintNFT         = "mint_nft"
	EventTypeBurnNFT         = "burn_nft"

	AttributeValueCategory = ModuleName

	AttributeKeySender         = "sender"
	AttributeKeyCreator        = "creator"
	AttributeKeyRecipient      = "recipient"
	AttributeKeyOwner          = "owner"
	AttributeKeyTokenID        = "token_id"
	AttributeKeyTokenURI       = "token_uri"
	AttributeKeyCollectionID   = "collection_id"
	AttributeKeyCollectionName = "collection_name"
)
