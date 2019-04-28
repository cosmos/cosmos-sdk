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

// RouterKey is nft
var RouterKey = "nft"

/* --------------------------------------------------------------------------- */
// MsgTransferNFT
/* --------------------------------------------------------------------------- */

// MsgTransferNFT defines a TransferNFT message
type MsgTransferNFT struct {
	Sender    sdk.AccAddress
	Recipient sdk.AccAddress
	Denom     Denom
	ID        TokenID
}

// NewMsgTransferNFT is a constructor function for MsgSetName
func NewMsgTransferNFT(sender, recipient sdk.AccAddress, denom Denom, id TokenID,
) MsgTransferNFT {
	return MsgTransferNFT{
		Sender:    sender,
		Recipient: recipient,
		Denom:     denom,
		ID:        id,
	}
}

// Route Implements Msg
func (msg MsgTransferNFT) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgTransferNFT) Type() string { return "transfer_nft" }

// ValidateBasic Implements Msg.
func (msg MsgTransferNFT) ValidateBasic() sdk.Error {
	if string(msg.Denom) == "" {
		return ErrInvalidCollection(DefaultCodespace)
	}
	if msg.ID.Empty() {
		return ErrInvalidNFT(DefaultCodespace)
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
	ID              TokenID
	Denom           Denom
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
func NewMsgEditNFTMetadata(owner sdk.AccAddress, denom Denom, id TokenID,
	editName, editDescription, editImage, editTokenURI bool,
	name, description, image, tokenURI string,
) MsgEditNFTMetadata {
	return MsgEditNFTMetadata{
		Owner:           owner,
		ID:              id,
		Denom:           Denom(strings.TrimSpace(string(denom))),
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
	if msg.ID.Empty() {
		return ErrInvalidNFT(DefaultCodespace)
	}
	if msg.Owner.Empty() {
		return sdk.ErrInvalidAddress("invalid owner address")
	}
	if !msg.EditName && !msg.EditDescription && !msg.EditImage && !msg.EditTokenURI {
		return ErrEmptyMetadata(DefaultCodespace)
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
	ID          TokenID
	Denom       Denom
	Name        string
	Description string
	Image       string
	TokenURI    string
}

// NewMsgMintNFT is a constructor function for MsgMintNFT
func NewMsgMintNFT(sender, recipient sdk.AccAddress, id TokenID, denom Denom, name string, description string, image string, tokenURI string,
) MsgMintNFT {
	return MsgMintNFT{
		Sender:      sender,
		Recipient:   recipient,
		ID:          id,
		Denom:       denom,
		Name:        name,
		Description: description,
		Image:       image,
		TokenURI:    tokenURI,
	}
}

// Route Implements Msg
func (msg MsgMintNFT) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgMintNFT) Type() string { return "mint_nft" }

// ValidateBasic Implements Msg.
func (msg MsgMintNFT) ValidateBasic() sdk.Error {
	if strings.TrimSpace(string(msg.Denom)) == "" {
		return ErrInvalidCollection(DefaultCodespace)
	}
	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress("invalid sender address")
	}
	if msg.ID.Empty() {
		return ErrInvalidNFT(DefaultCodespace)
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
	ID     TokenID
	Denom  Denom
}

// NewMsgBurnNFT is a constructor function for MsgBurnNFT
func NewMsgBurnNFT(sender sdk.AccAddress, id TokenID, denom Denom,
) MsgBurnNFT {
	return MsgBurnNFT{
		Sender: sender,
		ID:     id,
		Denom:  denom,
	}
}

// Route Implements Msg
func (msg MsgBurnNFT) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgBurnNFT) Type() string { return "burn_nft" }

// ValidateBasic Implements Msg.
func (msg MsgBurnNFT) ValidateBasic() sdk.Error {
	if strings.TrimSpace(string(msg.Denom)) == "" {
		return ErrInvalidCollection(DefaultCodespace)
	}
	if msg.ID.Empty() {
		return ErrInvalidNFT(DefaultCodespace)
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
	Denom  Denom
	ID     TokenID
}

// NewMsgBuyNFT is a constructor function for MsgBuyNFT
func NewMsgBuyNFT(sender, owner sdk.AccAddress, denom Denom, id TokenID,
) MsgBuyNFT {
	return MsgBuyNFT{
		Sender: sender,
		Denom:  denom,
		ID:     id,
	}
}

// Route Implements Msg
func (msg MsgBuyNFT) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgBuyNFT) Type() string { return "buy_nft" }

// ValidateBasic Implements Msg.
func (msg MsgBuyNFT) ValidateBasic() sdk.Error {
	if strings.TrimSpace(string(msg.Denom)) == "" {
		return ErrInvalidCollection(DefaultCodespace)
	}
	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress("invalid sender address")
	}
	if !msg.Amount.IsValid() {
		return sdk.ErrInvalidCoins("invalid amount provided")
	}
	if msg.ID.Empty() {
		return ErrInvalidNFT(DefaultCodespace)
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
