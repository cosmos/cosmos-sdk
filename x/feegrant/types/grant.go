package types

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ types.UnpackInterfacesMessage = &FeeAllowanceGrant{}
)

// ValidateBasic performs basic validation on
// FeeAllowanceGrant
func (a FeeAllowanceGrant) ValidateBasic() error {
	if a.Granter.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing granter address")
	}
	if a.Grantee.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing grantee address")
	}
	if a.Grantee.Equals(a.Granter) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "cannot self-grant fee authorization")
	}

	return a.GetFeeGrant().ValidateBasic()
}

func (a FeeAllowanceGrant) GetFeeGrant() FeeAllowanceI {
	allowance, ok := a.Allowance.GetCachedValue().(FeeAllowanceI)
	if !ok {
		return nil
	}

	return allowance
	// return a.Allowance.GetFeeAllowanceI()
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (a FeeAllowanceGrant) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var allowance FeeAllowanceI
	return unpacker.UnpackAny(a.Allowance, &allowance)
}

// PrepareForExport will m	ake all needed changes to the allowance to prepare to be
// re-imported at height 0, and return a copy of this grant.
func (a FeeAllowanceGrant) PrepareForExport(dumpTime time.Time, dumpHeight int64) FeeAllowanceGrant {
	err := a.GetFeeGrant().PrepareForExport(dumpTime, dumpHeight)
	if err != nil {
		return FeeAllowanceGrant{}
	}
	return a
}
