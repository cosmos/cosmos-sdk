package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant/exported"
)

// FeeAllowanceGrant is stored in the KVStore to record a grant with full context
type FeeAllowanceGrant struct {
	Granter   sdk.AccAddress        `json:"granter" yaml:"granter"`
	Grantee   sdk.AccAddress        `json:"grantee" yaml:"grantee"`
	Allowance exported.FeeAllowance `json:"allowance" yaml:"allowance"`
}

// ValidateBasic ensures that
func (a FeeAllowanceGrant) ValidateBasic() error {
	if a.Granter.Empty() {
		return sdk.ErrInvalidAddress("missing granter address")
	}
	if a.Grantee.Empty() {
		return sdk.ErrInvalidAddress("missing grantee address")
	}
	if a.Grantee.Equals(a.Granter) {
		return sdk.ErrInvalidAddress("cannot self-grant fees")
	}
	return a.Allowance.ValidateBasic()
}

// PrepareForExport will make all needed changes to the allowance to prepare to be
// re-imported at height 0, and return a copy of this grant.
func (a FeeAllowanceGrant) PrepareForExport(dumpTime time.Time, dumpHeight int64) FeeAllowanceGrant {
	a.Allowance = a.Allowance.PrepareForExport(dumpTime, dumpHeight)
	return a
}
