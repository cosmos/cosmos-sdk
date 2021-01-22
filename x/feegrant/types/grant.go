package types

import (
	fmt "fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	proto "github.com/gogo/protobuf/proto"
)

var (
	_ types.UnpackInterfacesMessage = &FeeAllowanceGrant{}
)

// NewFeeAllowanceGrant creates a new FeeAllowanceGrant.
//nolint:interfacer
func NewFeeAllowanceGrant(granter, grantee sdk.AccAddress, feeAllowance FeeAllowanceI) FeeAllowanceGrant {
	msg, ok := feeAllowance.(proto.Message)
	if !ok {
		panic(fmt.Errorf("cannot proto marshal %T", msg))
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		panic(err)
	}

	return FeeAllowanceGrant{
		Granter:   granter.String(),
		Grantee:   grantee.String(),
		Allowance: any,
	}
}

// ValidateBasic performs basic validation on
// FeeAllowanceGrant
func (a FeeAllowanceGrant) ValidateBasic() error {
	if a.Granter == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing granter address")
	}
	if a.Grantee == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing grantee address")
	}
	if a.Grantee == a.Granter {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "cannot self-grant fee authorization")
	}

	return a.GetFeeGrant().ValidateBasic()
}

// GetFeeGrant unpacks allowance
func (a FeeAllowanceGrant) GetFeeGrant() FeeAllowanceI {
	allowance, ok := a.Allowance.GetCachedValue().(FeeAllowanceI)
	if !ok {
		return nil
	}

	return allowance
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (a FeeAllowanceGrant) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var allowance FeeAllowanceI
	return unpacker.UnpackAny(a.Allowance, &allowance)
}

// PrepareForExport will make all needed changes to the allowance to prepare to be
// re-imported at height 0, and return a copy of this grant.
func (a FeeAllowanceGrant) PrepareForExport(dumpTime time.Time, dumpHeight int64) FeeAllowanceGrant {
	feegrant := a.GetFeeGrant().PrepareForExport(dumpTime, dumpHeight)
	if feegrant == nil {
		return FeeAllowanceGrant{}
	}

	granter, err := sdk.AccAddressFromBech32(a.Granter)
	if err != nil {
		return FeeAllowanceGrant{}
	}

	grantee, err := sdk.AccAddressFromBech32(a.Grantee)
	if err != nil {
		return FeeAllowanceGrant{}
	}

	return NewFeeAllowanceGrant(granter, grantee, feegrant)
}
