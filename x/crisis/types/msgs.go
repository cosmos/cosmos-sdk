package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ensure Msg interface compliance at compile time
var _ sdk.Msg = &MsgVerifyInvariant{}

// NewMsgVerifyInvariant creates a new MsgVerifyInvariant object
//
//nolint:interfacer
func NewMsgVerifyInvariant(sender sdk.AccAddress, invModeName, invRoute string) *MsgVerifyInvariant {
	return &MsgVerifyInvariant{
		Sender:              sender.String(),
		InvariantModuleName: invModeName,
		InvariantRoute:      invRoute,
	}
}

func (msg MsgVerifyInvariant) Route() string { return ModuleName }
func (msg MsgVerifyInvariant) Type() string  { return "verify_invariant" }

// get the bytes for the message signer to sign on
func (msg MsgVerifyInvariant) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{sender}
}

// GetSignBytes gets the sign bytes for the msg MsgVerifyInvariant
func (msg MsgVerifyInvariant) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
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
