package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

/* ---------------------------------------------------------------------------
MsgTransferNFT (nfts)
MsgEditNFTMetadata (nfts)
MsgMintNFT (mintable-nft)
MsgBurnNFT (burnable-nft)
MsgBuyNFT (nft-market)
--------------------------------------------------------------------------- */

// RouterKey is nfts
var RouterKey = "nfts"

/* --------------------------------------------------------------------------- */
// MsgTransferNFT
/* --------------------------------------------------------------------------- */

// MsgTransferNFT defines a TransferNFT message
type MsgTransferNFT struct {
	Sender    sdk.AccAddress
	Recipient sdk.AccAddress
	Denom     string
	ID        uint64
}

// NewMsgTransferNFT is a constructor function for MsgSetName
func NewMsgTransferNFT(sender, recipient sdk.AccAddress, denom string, id uint64) MsgTransferNFT {
	return MsgTransferNFT{
		Sender:    sender,
		Recipient: recipient,
		Denom:     strings.TrimSpace(denom),
		ID:        id,
	}
}

// Route Implements Msg
func (msg MsgTransferNFT) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgTransferNFT) Type() string { return "transfer_nft" }

// ValidateBasic Implements Msg.
func (msg MsgTransferNFT) ValidateBasic() sdk.Error {
	if msg.Denom == "" {
		return ErrInvalidCollection(DefaultCodespace)
	}
	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress("invalid sender address")
	}
	if msg.Recipient.Empty() {
		return sdk.ErrInvalidAddress("invalid recipient address")
	}

	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgTransferNFT) GetSignBytes() []byte {
	bz := cdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgTransferNFT) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

/* --------------------------------------------------------------------------- */
// MsgEditNFTMetadata
/* --------------------------------------------------------------------------- */

// MsgEditNFTMetadata edits an NFT's metadata
type MsgEditNFTMetadata struct {
	Owner           sdk.AccAddress
	ID              uint64
	Denom           string
	EditName        bool
	EditDescription bool
	EditImage       bool
	EditTokenURI    bool
	Name            string
	Description     string
	Image           string
	TokenURI        string
}

// NewMsgEditNFTMetadata is a constructor function for MsgSetName
func NewMsgEditNFTMetadata(owner sdk.AccAddress, id uint64,
	editName, editDescription, editImage, editTokenURI bool,
	denom, name, description, image, tokenURI string,
) MsgEditNFTMetadata {
	return MsgEditNFTMetadata{
		Owner:           owner,
		ID:              id,
		Denom:           strings.TrimSpace(denom),
		EditName:        editName,
		EditDescription: editDescription,
		EditImage:       editImage,
		EditTokenURI:    editTokenURI,
		Name:            strings.TrimSpace(name),
		Description:     strings.TrimSpace(description),
		Image:           strings.TrimSpace(image),
		TokenURI:        strings.TrimSpace(tokenURI),
	}
}

// Route Implements Msg
func (msg MsgEditNFTMetadata) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgEditNFTMetadata) Type() string { return "edit_metadata" }

// ValidateBasic Implements Msg.
func (msg MsgEditNFTMetadata) ValidateBasic() sdk.Error {
	if msg.Owner.Empty() {
		return sdk.ErrInvalidAddress("invalid owner address")
	}
	if !msg.EditName && !msg.EditDescription && !msg.EditImage && !msg.EditTokenURI {
		return ErrEmptyMetadata(DefaultCodespace, "")
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgEditNFTMetadata) GetSignBytes() []byte {
	bz := cdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgEditNFTMetadata) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}

/* --------------------------------------------------------------------------- */
// MsgMintNFT
/* --------------------------------------------------------------------------- */

// MsgMintNFT defines a MintNFT message
type MsgMintNFT struct {
	Sender      sdk.AccAddress
	Recipient   sdk.AccAddress
	ID          uint64
	Denom       string
	Name        string
	Description string
	Image       string
	TokenURI    string
}

// NewMsgMintNFT is a constructor function for MsgMintNFT
func NewMsgMintNFT(sender, recipient sdk.AccAddress, id uint64, denom, name, description, image, tokenURI string) MsgMintNFT {
	return MsgMintNFT{
		Sender:      sender,
		Recipient:   recipient,
		ID:          id,
		Denom:       strings.TrimSpace(denom),
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		Image:       strings.TrimSpace(image),
		TokenURI:    strings.TrimSpace(tokenURI),
	}
}

// Route Implements Msg
func (msg MsgMintNFT) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgMintNFT) Type() string { return "mint_nft" }

// ValidateBasic Implements Msg.
func (msg MsgMintNFT) ValidateBasic() sdk.Error {
	if msg.Denom == "" {
		return ErrInvalidCollection(DefaultCodespace)
	}
	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress("invalid sender address")
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgMintNFT) GetSignBytes() []byte {
	bz := cdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgMintNFT) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

/* --------------------------------------------------------------------------- */
// MsgBurnNFT
/* --------------------------------------------------------------------------- */

// MsgBurnNFT defines a BurnNFT message
type MsgBurnNFT struct {
	Sender sdk.AccAddress
	ID     uint64
	Denom  string
}

// NewMsgBurnNFT is a constructor function for MsgBurnNFT
func NewMsgBurnNFT(sender sdk.AccAddress, id uint64, denom string) MsgBurnNFT {
	return MsgBurnNFT{
		Sender: sender,
		ID:     id,
		Denom:  strings.TrimSpace(denom),
	}
}

// Route Implements Msg
func (msg MsgBurnNFT) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgBurnNFT) Type() string { return "burn_nft" }

// ValidateBasic Implements Msg.
func (msg MsgBurnNFT) ValidateBasic() sdk.Error {
	if msg.Denom == "" {
		return ErrInvalidCollection(DefaultCodespace)
	}
	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress("invalid sender address")
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgBurnNFT) GetSignBytes() []byte {
	bz := cdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgBurnNFT) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

/* --------------------------------------------------------------------------- */
// MsgBuyNFT
/* --------------------------------------------------------------------------- */

// MsgBuyNFT defines a MsgBuyNFT message
type MsgBuyNFT struct {
	Sender sdk.AccAddress
	Amount sdk.Coins
	Denom  string
	ID     uint64
}

// NewMsgBuyNFT is a constructor function for MsgBuyNFT
func NewMsgBuyNFT(sender, owner sdk.AccAddress, denom string, id uint64) MsgBuyNFT {
	return MsgBuyNFT{
		Sender: sender,
		Denom:  strings.TrimSpace(denom),
		ID:     id,
	}
}

// Route Implements Msg
func (msg MsgBuyNFT) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgBuyNFT) Type() string { return "buy_nft" }

// ValidateBasic Implements Msg.
func (msg MsgBuyNFT) ValidateBasic() sdk.Error {
	if msg.Denom == "" {
		return ErrInvalidCollection(DefaultCodespace)
	}
	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress("invalid sender address")
	}
	if !msg.Amount.IsValid() {
		return sdk.ErrInvalidCoins("invalid amount provided")
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgBuyNFT) GetSignBytes() []byte {
	bz := cdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgBuyNFT) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
