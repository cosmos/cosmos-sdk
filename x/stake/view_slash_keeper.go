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
func (v ViewSlashKeeper) GetDelegatorBond(ctx sdk.Context,
	delegatorAddr, candidateAddr sdk.Address) (bond DelegatorBond, found bool) {
	return v.keeper.GetDelegatorBond(ctx, delegatorAddr, candidateAddr)
}

// load n delegator bonds
func (v ViewSlashKeeper) GetDelegatorBonds(ctx sdk.Context,
	delegator sdk.Address, maxRetrieve int16) (bonds []DelegatorBond) {
	return v.keeper.GetDelegatorBonds(ctx, delegator, maxRetrieve)
}
