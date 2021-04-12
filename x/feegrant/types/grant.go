package types

import (
	"time"

	proto "github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ types.UnpackInterfacesMessage = &FeeAllowanceGrant{}
)

// NewFeeAllowanceGrant creates a new FeeAllowanceGrant.
//nolint:interfacer
func NewFeeAllowanceGrant(granter, grantee sdk.AccAddress, feeAllowance FeeAllowanceI) (FeeAllowanceGrant, error) {
	msg, ok := feeAllowance.(proto.Message)
	if !ok {
		return FeeAllowanceGrant{}, sdkerrors.Wrapf(sdkerrors.ErrPackAny, "cannot proto marshal %T", feeAllowance)
	}

	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return FeeAllowanceGrant{}, err
	}

	return FeeAllowanceGrant{
		Granter:   granter.String(),
		Grantee:   grantee.String(),
		Allowance: any,
	}, nil
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

	f, err := a.GetFeeGrant()
	if err != nil {
		return err
	}

	return f.ValidateBasic()
}

// GetFeeGrant unpacks allowance
func (a FeeAllowanceGrant) GetFeeGrant() (FeeAllowanceI, error) {
	allowance, ok := a.Allowance.GetCachedValue().(FeeAllowanceI)
	if !ok {
		return nil, sdkerrors.Wrap(ErrNoAllowance, "failed to get allowance")
	}

	return allowance, nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (a FeeAllowanceGrant) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var allowance FeeAllowanceI
	return unpacker.UnpackAny(a.Allowance, &allowance)
}

// PrepareForExport will make all needed changes to the allowance to prepare to be
// re-imported at height 0, and return a copy of this grant.
func (a FeeAllowanceGrant) PrepareForExport(dumpTime time.Time, dumpHeight int64) FeeAllowanceGrant {
	f, err := a.GetFeeGrant()
	if err != nil {
		return FeeAllowanceGrant{}
	}

	feegrant := f.PrepareForExport(dumpTime, dumpHeight)
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

	grant, err := NewFeeAllowanceGrant(granter, grantee, feegrant)
	if err != nil {
		return FeeAllowanceGrant{}
	}

	return grant
}
