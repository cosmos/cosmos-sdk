package nft

import (
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
)

const (
	// TypeMsgSend nft message types
	TypeMsgSend = "send"
)

var (
	_ sdk.Msg = &MsgSend{}
	// For amino support.
	_ legacytx.LegacyMsg = &MsgSend{}
)

// GetSigners implements the Msg.ValidateBasic method.
func (m MsgSend) ValidateBasic() error {
	if err := ValidateClassID(m.ClassID); err != nil {
		return sdkerrors.Wrap(ErrInvalidID, "class id")
	}

	if err := ValidateNFTID(m.ID); err != nil {
		return sdkerrors.Wrap(ErrInvalidID, "class id")
	}

	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	_, err = sdk.AccAddressFromBech32(m.Receiver)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid receiver address")
	}
	return nil
}

// GetSigners implements the Msg.GetSigners method.
func (m MsgSend) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// Type implements the LegacyMsg.Type method.
func (m MsgSend) Type() string {
	return sdk.MsgTypeURL(&m)
}

// Route implements the LegacyMsg.Route method.
func (m MsgSend) Route() string {
	return sdk.MsgTypeURL(&m)
}

// GetSignBytes implements the LegacyMsg.GetSignBytes method.
func (m MsgSend) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&m))
}
