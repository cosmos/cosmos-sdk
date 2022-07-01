package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// distribution message types
const (
	TypeMsgSetAutoDelegation   = "set_auto_delegation"
	TypeMsgUnSetAutoDelegation = "unset_auto_delegation"
)

// Verify interface at compile time
var _, _ sdk.Msg = &MsgSetAutoDelegation{}, &MsgUnSetAutoDelegation{}

func NewMsgSetAutoDelegation(delAddr sdk.AccAddress, minBalance sdk.Coins) *MsgSetAutoDelegation {
	return &MsgSetAutoDelegation{
		DelegatorAddress: delAddr.String(),
		MinBalance:       minBalance,
	}
}

func (msg MsgSetAutoDelegation) Route() string { return ModuleName }
func (msg MsgSetAutoDelegation) Type() string  { return TypeMsgSetAutoDelegation }

// Return address that must sign over msg.GetSignBytes()
func (msg MsgSetAutoDelegation) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	return []sdk.AccAddress{delegator}
}

// get the bytes for the message signer to sign on
func (msg MsgSetAutoDelegation) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgSetAutoDelegation) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DelegatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}
	return nil
}

/*--------------------*/

func NewMsgUnSetAutoDelegation(delAddr sdk.AccAddress) *MsgUnSetAutoDelegation {
	return &MsgUnSetAutoDelegation{
		DelegatorAddress: delAddr.String(),
	}
}

func (msg MsgUnSetAutoDelegation) Route() string { return ModuleName }
func (msg MsgUnSetAutoDelegation) Type() string  { return TypeMsgUnSetAutoDelegation }

// Return address that must sign over msg.GetSignBytes()
func (msg MsgUnSetAutoDelegation) GetSigners() []sdk.AccAddress {
	delegator, _ := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	return []sdk.AccAddress{delegator}
}

// get the bytes for the message signer to sign on
func (msg MsgUnSetAutoDelegation) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgUnSetAutoDelegation) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DelegatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}
	return nil
}
