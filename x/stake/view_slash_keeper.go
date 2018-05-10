package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// keeper to view information & slash validators
// will be used by governance module
type ViewSlashKeeper struct {
	keeper Keeper
}

// NewViewSlashKeeper creates a keeper restricted to
// viewing information & slashing validators
func NewViewSlashKeeper(k Keeper) ViewSlashKeeper {
	return ViewSlashKeeper{k}
}

// load a delegator bond
func (v ViewSlashKeeper) GetDelegation(ctx sdk.Context,
	delegatorAddr, validatorAddr sdk.Address) (bond Delegation, found bool) {
	return v.keeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
}

// load n delegator bonds
func (v ViewSlashKeeper) GetDelegations(ctx sdk.Context,
	delegator sdk.Address, maxRetrieve int16) (bonds []Delegation) {
	return v.keeper.GetDelegations(ctx, delegator, maxRetrieve)
}
