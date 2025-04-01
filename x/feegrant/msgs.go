package feegrant

import (
	"github.com/cosmos/gogoproto/proto"

	errorsmod "cosmossdk.io/errors"

	codectypes "github.com/cosmos/gogoproto/types/any"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_, _ sdk.Msg                            = &MsgGrantAllowance{}, &MsgRevokeAllowance{}
	_    codectypes.UnpackInterfacesMessage = &MsgGrantAllowance{}
)

// NewMsgGrantAllowance creates a new MsgGrantAllowance.
func NewMsgGrantAllowance(feeAllowance FeeAllowanceI, granter, grantee sdk.AccAddress) (*MsgGrantAllowance, error) {
	msg, ok := feeAllowance.(proto.Message)
	if !ok {
		return nil, errorsmod.Wrapf(sdkerrors.ErrPackAny, "cannot proto marshal %T", msg)
	}
	value, err := codectypes.NewAnyWithCacheWithValue(msg)
	if err != nil {
		return nil, err
	}

	return &MsgGrantAllowance{
		Granter:   granter.String(),
		Grantee:   grantee.String(),
		Allowance: value,
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
func (msg MsgGrantAllowance) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var allowance FeeAllowanceI
	return unpacker.UnpackAny(msg.Allowance, &allowance)
}

// NewMsgRevokeAllowance returns a message to revoke a fee allowance for a given
// granter and grantee
func NewMsgRevokeAllowance(granter, grantee sdk.AccAddress) MsgRevokeAllowance {
	return MsgRevokeAllowance{Granter: granter.String(), Grantee: grantee.String()}
}
