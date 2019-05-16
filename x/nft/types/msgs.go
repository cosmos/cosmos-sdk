package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
	Owner       sdk.AccAddress
	ID          uint64
	Denom       string
	Name        string
	Description string
	Image       string
	TokenURI    string
}

// NewMsgEditNFTMetadata is a constructor function for MsgSetName
func NewMsgEditNFTMetadata(owner sdk.AccAddress, id uint64,
	denom, name, description, image, tokenURI string,
) MsgEditNFTMetadata {
	return MsgEditNFTMetadata{
		Owner:       owner,
		ID:          id,
		Denom:       strings.TrimSpace(denom),
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		Image:       strings.TrimSpace(image),
		TokenURI:    strings.TrimSpace(tokenURI),
	}
}

// Route Implements Msg
func (msg MsgEditNFTMetadata) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgEditNFTMetadata) Type() string { return "edit_nft_metadata" }

// ValidateBasic Implements Msg.
func (msg MsgEditNFTMetadata) ValidateBasic() sdk.Error {
	if msg.Owner.Empty() {
		return sdk.ErrInvalidAddress("invalid owner address")
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
