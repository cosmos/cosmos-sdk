package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/fee_grant/exported"
)

// MsgDelegateFeeAllowance adds permission for Grantee to spend up to Allowance
// of fees from the account of Granter.
// If there was already an existing grant, this overwrites it.
type MsgDelegateFeeAllowance struct {
	Granter   sdk.AccAddress        `json:"granter" yaml:"granter"`
	Grantee   sdk.AccAddress        `json:"grantee" yaml:"grantee"`
	Allowance exported.FeeAllowance `json:"allowance" yaml:"allowance"`
}

func NewMsgDelegateFeeAllowance(granter sdk.AccAddress, grantee sdk.AccAddress, allowance exported.FeeAllowance) MsgDelegateFeeAllowance {
	return MsgDelegateFeeAllowance{Granter: granter, Grantee: grantee, Allowance: allowance}
}

func (msg MsgDelegateFeeAllowance) Route() string {
	return RouterKey
}

func (msg MsgDelegateFeeAllowance) Type() string {
	return "delegate-fee-allowance"
}

func (msg MsgDelegateFeeAllowance) ValidateBasic() sdk.Error {
	if msg.Granter.Empty() {
		return sdk.ErrInvalidAddress("missing granter address")
	}
	if msg.Grantee.Empty() {
		return sdk.ErrInvalidAddress("missing grantee address")
	}
	return sdk.ConvertError(msg.Allowance.ValidateBasic())
}

func (msg MsgDelegateFeeAllowance) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgDelegateFeeAllowance) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Granter}
}

// MsgRevokeFeeAllowance removes any existing FeeAllowance from Granter to Grantee.
type MsgRevokeFeeAllowance struct {
	Granter sdk.AccAddress `json:"granter" yaml:"granter"`
	Grantee sdk.AccAddress `json:"grantee" yaml:"granter"`
}

func NewMsgRevokeFeeAllowance(granter sdk.AccAddress, grantee sdk.AccAddress) MsgRevokeFeeAllowance {
	return MsgRevokeFeeAllowance{Granter: granter, Grantee: grantee}
}

func (msg MsgRevokeFeeAllowance) Route() string {
	return RouterKey
}

func (msg MsgRevokeFeeAllowance) Type() string {
	return "revoke-fee-allowance"
}

func (msg MsgRevokeFeeAllowance) ValidateBasic() sdk.Error {
	if msg.Granter.Empty() {
		return sdk.ErrInvalidAddress("missing granter address")
	}
	if msg.Grantee.Empty() {
		return sdk.ErrInvalidAddress("missing grantee address")
	}
	return nil
}

func (msg MsgRevokeFeeAllowance) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgRevokeFeeAllowance) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Granter}
}
