package types

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

// ensure Msg interface compliance at compile time
var (
	_, _ sdk.Msg            = &MsgVerifyInvariant{}, &MsgUpdateParams{}
	_, _ legacytx.LegacyMsg = &MsgVerifyInvariant{}, &MsgUpdateParams{}
)

// NewMsgVerifyInvariant creates a new MsgVerifyInvariant object
func NewMsgVerifyInvariant(sender sdk.AccAddress, invModeName, invRoute string) *MsgVerifyInvariant {
	return &MsgVerifyInvariant{
		Sender:              sender.String(),
		InvariantModuleName: invModeName,
		InvariantRoute:      invRoute,
	}
}

// get the bytes for the message signer to sign on
func (msg MsgVerifyInvariant) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{sender}
}

// GetSignBytes gets the sign bytes for the msg MsgVerifyInvariant
func (msg MsgVerifyInvariant) GetSignBytes() []byte {
	bz := aminoCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgVerifyInvariant) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}
	return nil
}

// FullInvariantRoute - get the messages full invariant route
func (msg MsgVerifyInvariant) FullInvariantRoute() string {
	return msg.InvariantModuleName + "/" + msg.InvariantRoute
}

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// GetSignBytes returns the raw bytes for a MsgUpdateParams message that
// the expected signer needs to sign.
func (msg MsgUpdateParams) GetSignBytes() []byte {
	bz := aminoCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic performs basic MsgUpdateParams message validation.
func (msg MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if !msg.ConstantFee.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, "invalid costant fee")
	}

	if msg.ConstantFee.IsNegative() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, "negative costant fee")
	}

	return nil
}
