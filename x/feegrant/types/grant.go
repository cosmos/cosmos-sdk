package types

import (
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
	return a.Allowance.GetFeeAllowanceI()
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
