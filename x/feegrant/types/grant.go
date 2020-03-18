package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewFeeAllowanceGrantBase(granter sdk.AccAddress, grantee sdk.AccAddress) FeeAllowanceGrantBase {
	return FeeAllowanceGrantBase{
		Granter: granter,
		Grantee: grantee,
	}
}

////TODO these should be implemented in std/msgs.go
//// ValidateBasic ensures that
//func (a FeeAllowanceGrantBase) ValidateBasic() error {
//	if a.Granter.Empty() {
//		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing granter address")
//	}
//	if a.Grantee.Empty() {
//		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing grantee address")
//	}
//	if a.Grantee.Equals(a.Granter) {
//		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "cannot self-grant fee authorization")
//	}
//
//	return a.GetAllowance().GetFeeAllowance().ValidateBasic()
//}
//
//// PrepareForExport will make all needed changes to the allowance to prepare to be
//// re-imported at height 0, and return a copy of this grant.
//func (a FeeAllowanceGrantBase) PrepareForExport(dumpTime time.Time, dumpHeight int64) exported.FeeAllowanceGrant {
//	err := a.GetAllowance().SetFeeAllowance(a.GetAllowance().GetFeeAllowance().PrepareForExport(dumpTime, dumpHeight))
//	if err != nil {
//		//TODO handle this error
//	}
//	return a
//}
