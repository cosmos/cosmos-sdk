package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/feegrant/exported"
)

type MsgGrantFeeAllowance struct {
	Granter   sdk.AccAddress        `json:"granter" yaml:"granter"`
	Grantee   sdk.AccAddress        `json:"grantee" yaml:"grantee"`
	Allowance exported.FeeAllowance `json:"allowance" yaml:"allowance"`
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

func (msg MsgRevokeFeeAllowance) ValidateBasic() error {
	if msg.Granter.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing granter address")
	}
	if msg.Grantee.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing grantee address")
	}

	return nil
}

func (msg MsgRevokeFeeAllowance) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgRevokeFeeAllowance) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Granter}
}
