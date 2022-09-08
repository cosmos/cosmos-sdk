package feegrant

import (
	"github.com/cosmos/gogoproto/proto"
	gogoprotoany "github.com/cosmos/gogoproto/types/any"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_, _ sdk.Msg                              = &MsgGrantAllowance{}, &MsgRevokeAllowance{}
	_    gogoprotoany.UnpackInterfacesMessage = &MsgGrantAllowance{}
)

// NewMsgGrantAllowance creates a new MsgGrantAllowance.
//
//nolint:interfacer
func NewMsgGrantAllowance(feeAllowance FeeAllowanceI, granter, grantee sdk.AccAddress) (*MsgGrantAllowance, error) {
	msg, ok := feeAllowance.(proto.Message)
	if !ok {
		return nil, errorsmod.Wrapf(sdkerrors.ErrPackAny, "cannot proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return &MsgGrantAllowance{
		Granter:   granter,
		Grantee:   grantee,
		Allowance: any,
	}, nil
}

// GetFeeAllowanceI returns unpacked FeeAllowance
func (msg MsgGrantAllowance) GetFeeAllowanceI() (FeeAllowanceI, error) {
	allowance, ok := msg.Allowance.GetCachedValue().(FeeAllowanceI)
	if !ok {
		return nil, errorsmod.Wrap(ErrNoAllowance, "failed to get allowance")
	}

	return allowance, nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgGrantAllowance) UnpackInterfaces(unpacker gogoprotoany.AnyUnpacker) error {
	var allowance FeeAllowanceI
	return unpacker.UnpackAny(msg.Allowance, &allowance)
}

// NewMsgRevokeAllowance returns a message to revoke a fee allowance for a given
// granter and grantee
//
//nolint:interfacer
func NewMsgRevokeAllowance(granter sdk.AccAddress, grantee sdk.AccAddress) MsgRevokeAllowance {
	return MsgRevokeAllowance{Granter: granter.String(), Grantee: grantee.String()}
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgRevokeAllowance) ValidateBasic() error {
	if msg.Granter == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing granter address")
	}
	if msg.Grantee == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing grantee address")
	}
	if msg.Grantee == msg.Granter {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "addresses must be different")
	}

	return nil
}

// GetSigners gets the granter address associated with an Allowance
// to revoke.
func (msg MsgRevokeAllowance) GetSigners() []sdk.AccAddress {
	granter, err := sdk.AccAddressFromBech32(msg.Granter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{granter}
}

// Type implements the LegacyMsg.Type method.
func (msg MsgRevokeAllowance) Type() string {
	return sdk.MsgTypeURL(&msg)
}

// Route implements the LegacyMsg.Route method.
func (msg MsgRevokeAllowance) Route() string {
	return sdk.MsgTypeURL(&msg)
}

// GetSignBytes implements the LegacyMsg.GetSignBytes method.
func (msg MsgRevokeAllowance) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&msg))
}
