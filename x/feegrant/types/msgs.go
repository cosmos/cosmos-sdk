package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (msg MsgGrantFeeAllowance) NewMsgGrantFeeAllowance(feeAllowance *FeeAllowanceI, granter, grantee sdk.AccAddress) (MsgGrantFeeAllowance, error) {
	return MsgGrantFeeAllowance{
		Granter:   granter,
		Grantee:   grantee,
		Allowance: feeAllowance,
	}, nil
}

func (msg MsgGrantFeeAllowance) GetFeeGrant() FeeAllowanceI {
	return msg.Allowance.GetFeeAllowanceI()
}

func (msg MsgGrantFeeAllowance) Route() string {
	return RouterKey
}

func (msg MsgGrantFeeAllowance) Type() string {
	return "grant-fee-allowance"
}

func (msg MsgGrantFeeAllowance) ValidateBasic() error {
	if msg.Granter.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing granter address")
	}
	if msg.Grantee.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing grantee address")
	}

	return nil
}

func (msg MsgGrantFeeAllowance) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgGrantFeeAllowance) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Granter}
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
