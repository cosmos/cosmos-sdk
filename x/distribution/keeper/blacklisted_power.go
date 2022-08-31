package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// get the delegator withdraw address, defaulting to the delegator address
func (k Keeper) GetBlacklistedPower(ctx sdk.Context, blockHeight string) (val types.BlacklistedPower, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get([]byte(blockHeight))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// set the blacklisted power entry
func (k Keeper) SetBlacklistedPower(ctx sdk.Context, blockHeight string, blacklistedPower types.BlacklistedPower) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&blacklistedPower)
	store.Set(types.GetBlacklistedPowerKey(blockHeight), b)
}

// delete a blacklisted power entry
func (k Keeper) DeleteBlacklistedPower(ctx sdk.Context, blockHeight string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetBlacklistedPowerKey(blockHeight))
}
