package v1

import (
	"errors"

	"github.com/cosmos/gogoproto/proto"
	gogoprotoany "github.com/cosmos/gogoproto/types/any"

	errorsmod "cosmossdk.io/errors"
	xfeegrant "cosmossdk.io/x/feegrant"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewGrant constructor
func NewGrant(feeAllowance xfeegrant.FeeAllowanceI) (*Grant, error) {
	msg, ok := feeAllowance.(proto.Message)
	if !ok {
		return nil, errorsmod.Wrapf(sdkerrors.ErrPackAny, "cannot proto marshal %T", msg)
	}
	value, err := types.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return &Grant{Allowance: value}, nil
}

// GetFeeAllowanceI returns unpacked FeeAllowance
func (msg Grant) GetFeeAllowanceI() (xfeegrant.FeeAllowanceI, error) {
	allowance, ok := msg.Allowance.GetCachedValue().(xfeegrant.FeeAllowanceI)
	if !ok {
		return nil, errors.New("failed to get allowance")
	}

	return allowance, nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg Grant) UnpackInterfaces(unpacker gogoprotoany.AnyUnpacker) error {
	var allowance xfeegrant.FeeAllowanceI
	return unpacker.UnpackAny(msg.Allowance, &allowance)
}
