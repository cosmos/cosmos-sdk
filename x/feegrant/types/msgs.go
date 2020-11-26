package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
)

var (
	_, _ sdk.Msg                       = &MsgGrantFeeAllowance{}, &MsgRevokeFeeAllowance{}
	_    types.UnpackInterfacesMessage = &MsgGrantFeeAllowance{}
)

// feegrant message types
const (
	TypeMsgGrantFeeAllowance  = "grant_fee_allowance"
	TypeMsgRevokeFeeAllowance = "revoke_fee_allowance"
)

// NewMsgGrantFeeAllowance creates a new MsgGrantFeeAllowance.
//nolint:interfacer
func (msg MsgGrantFeeAllowance) NewMsgGrantFeeAllowance(feeAllowance FeeAllowanceI, granter, grantee sdk.AccAddress) (*MsgGrantFeeAllowance, error) {
	msg1, ok := feeAllowance.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("cannot proto marshal %T", msg1)
	}
	any, err := types.NewAnyWithValue(msg1)
	if err != nil {
		return nil, err
	}

	return &MsgGrantFeeAllowance{
		Granter:   granter,
		Grantee:   grantee,
		Allowance: any,
	}, nil
}

// func (msg MsgGrantFeeAllowance) GetFeeGrant() FeeAllowanceI {
// 	return msg.GetFeeAllowanceI()
// }

// Route implements the sdk.Msg interface.
func (msg MsgGrantFeeAllowance) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg MsgGrantFeeAllowance) Type() string {
	return TypeMsgGrantFeeAllowance
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgGrantFeeAllowance) ValidateBasic() error {
	if msg.Granter.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing granter address")
	}
	if msg.Grantee.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing grantee address")
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgGrantFeeAllowance) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgGrantFeeAllowance) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Granter}
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
func (msg MsgGrantFeeAllowance) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var allowance FeeAllowanceI
	return unpacker.UnpackAny(msg.Allowance, &allowance)
}

func NewMsgRevokeFeeAllowance(granter sdk.AccAddress, grantee sdk.AccAddress) MsgRevokeFeeAllowance {
	return MsgRevokeFeeAllowance{Granter: granter, Grantee: grantee}
}

func (msg MsgRevokeFeeAllowance) Route() string {
	return RouterKey
}

func (msg MsgRevokeFeeAllowance) Type() string {
	return TypeMsgRevokeFeeAllowance
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
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgRevokeFeeAllowance) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Granter}
}
