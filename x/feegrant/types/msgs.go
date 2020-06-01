package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
)

func NewMsgGrantFeeAllowance(granter, grantee sdk.AccAddress, feeAllowance FeeAllowanceI) (*MsgGrantFeeAllowance, error) {
	m := &MsgGrantFeeAllowance{
		Granter: granter,
		Grantee: grantee,
	}

	err := m.SetAllowance(feeAllowance)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (msg MsgGrantFeeAllowance) GetFeeGrant() FeeAllowanceI {
	feeAllowance, ok := msg.Allowance.GetCachedValue().(FeeAllowanceI)
	if !ok {
		return nil
	}
	return feeAllowance
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

func (m *MsgGrantFeeAllowance) SetAllowance(FeeAllowanceI interface{}) error {
	a, ok := FeeAllowanceI.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", a)
	}
	any, err := types.NewAnyWithValue(a)
	if err != nil {
		return err
	}
	m.Allowance = any
	return nil
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
