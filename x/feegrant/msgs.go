package feegrant

import (
	"github.com/cosmos/gogoproto/proto"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

var (
	_, _ sdk.Msg = &MsgGrantAllowance{}, &MsgRevokeAllowance{}
	// For amino support.
	_, _ legacytx.LegacyMsg = &MsgGrantAllowance{}, &MsgRevokeAllowance{}

	_ types.UnpackInterfacesMessage = &MsgGrantAllowance{}
)

// NewMsgGrantAllowance creates a new MsgGrantAllowance.
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
		Granter:   granter.String(),
		Grantee:   grantee.String(),
		Allowance: any,
	}, nil
}

// GetSigners gets the granter account associated with an allowance
func (msg MsgGrantAllowance) GetSigners() []sdk.AccAddress {
	granter, _ := sdk.AccAddressFromBech32(msg.Granter)
	return []sdk.AccAddress{granter}
}

// GetSignBytes implements the LegacyMsg.GetSignBytes method.
func (msg MsgGrantAllowance) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
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
func (msg MsgGrantAllowance) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var allowance FeeAllowanceI
	return unpacker.UnpackAny(msg.Allowance, &allowance)
}

// NewMsgRevokeAllowance returns a message to revoke a fee allowance for a given
// granter and grantee
//

func NewMsgRevokeAllowance(granter, grantee sdk.AccAddress) MsgRevokeAllowance {
	return MsgRevokeAllowance{Granter: granter.String(), Grantee: grantee.String()}
}

// GetSigners gets the granter address associated with an Allowance
// to revoke.
func (msg MsgRevokeAllowance) GetSigners() []sdk.AccAddress {
	granter, _ := sdk.AccAddressFromBech32(msg.Granter)
	return []sdk.AccAddress{granter}
}

// GetSignBytes implements the LegacyMsg.GetSignBytes method.
func (msg MsgRevokeAllowance) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}
