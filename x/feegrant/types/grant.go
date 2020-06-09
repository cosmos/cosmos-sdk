package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewFeeAllowanceGrant(granter, grantee sdk.AccAddress, allowance FeeAllowanceI) FeeAllowanceGrant {
	a := &FeeAllowanceGrant{
		Granter: granter,
		Grantee: grantee,
	}

	err := a.SetAllowance(allowance)
	if err != nil {
		return FeeAllowanceGrant{}
	}

	return *a
}

// ValidateBasic performs basic validation on
// FeeAllowanceGrant
func (a *FeeAllowanceGrant) ValidateBasic() error {
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

func (a *FeeAllowanceGrant) GetFeeGrant() FeeAllowanceI {
	fmt.Printf("this is a %v", a.Allowance.GetCachedValue())
	feeAllowance, ok := a.Allowance.GetCachedValue().(FeeAllowanceI)
	if !ok {
		return nil
	}

	return feeAllowance
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

func (a *FeeAllowanceGrant) SetAllowance(FeeAllowanceI interface{}) error {
	allowance, ok := FeeAllowanceI.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", a)
	}
	any, err := types.NewAnyWithValue(allowance)
	if err != nil {
		return err
	}
	a.Allowance = any
	return nil
}
