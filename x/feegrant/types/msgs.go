package types

import (
	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_, _ sdk.MsgRequest                = &MsgGrantFeeAllowance{}, &MsgRevokeFeeAllowance{}
	_    types.UnpackInterfacesMessage = &MsgGrantFeeAllowance{}
)

// feegrant message types
const (
	TypeMsgGrantFeeAllowance  = "grant_fee_allowance"
	TypeMsgRevokeFeeAllowance = "revoke_fee_allowance"
)

// NewMsgGrantFeeAllowance creates a new MsgGrantFeeAllowance.
//nolint:interfacer
func NewMsgGrantFeeAllowance(feeAllowance FeeAllowanceI, granter, grantee sdk.AccAddress) (*MsgGrantFeeAllowance, error) {
	msg, ok := feeAllowance.(proto.Message)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrPackAny, "cannot proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return &MsgGrantFeeAllowance{
		Granter:   granter.String(),
		Grantee:   grantee.String(),
		Allowance: any,
	}, nil
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgGrantFeeAllowance) ValidateBasic() error {
	if msg.Granter == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing granter address")
	}
	if msg.Grantee == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing grantee address")
	}
	if msg.Grantee == msg.Granter {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "cannot self-grant fee authorization")
	}

	return msg.GetFeeAllowanceI().ValidateBasic()
}

func (msg MsgGrantFeeAllowance) GetSigners() []sdk.AccAddress {
	granter, err := sdk.AccAddressFromBech32(msg.Granter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{granter}
}

// GetFeeAllowanceI returns unpacked FeeAllowance
func (msg MsgGrantFeeAllowance) GetFeeAllowanceI() FeeAllowanceI {
	allowance, ok := msg.Allowance.GetCachedValue().(FeeAllowanceI)
	if !ok {
		return nil
	}

	return allowance
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgGrantFeeAllowance) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var allowance FeeAllowanceI
	return unpacker.UnpackAny(msg.Allowance, &allowance)
}

//nolint:interfacer
func NewMsgRevokeFeeAllowance(granter sdk.AccAddress, grantee sdk.AccAddress) MsgRevokeFeeAllowance {
	return MsgRevokeFeeAllowance{Granter: granter.String(), Grantee: grantee.String()}
}

func (msg MsgRevokeFeeAllowance) ValidateBasic() error {
	if msg.Granter == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing granter address")
	}
	if msg.Grantee == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing grantee address")
	}

	return nil
}

func (msg MsgRevokeFeeAllowance) GetSigners() []sdk.AccAddress {
	granter, err := sdk.AccAddressFromBech32(msg.Granter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{granter}
}
