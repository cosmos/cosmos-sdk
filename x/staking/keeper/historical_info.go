package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetHistoricalInfo gets the historical info at a given height
func (k Keeper) GetHistoricalInfo(ctx sdk.Context, height int64) (hi types.HistoricalInfo, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetHistoricalInfoKey(height)

	value := store.Get(key)

	if value == nil {
		return types.HistoricalInfo{}, false
	}

	hi = types.MustUnmarshalHistoricalInfo(k.cdc, value)
	return hi, true
}

// SetHistoricalInfo sets the historical info at a given height
func (k Keeper) SetHistoricalInfo(ctx sdk.Context, height int64, hi types.HistoricalInfo) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetHistoricalInfoKey(height)

	value := types.MustMarshalHistoricalInfo(k.cdc, hi)
	store.Set(key, value)
}

// DeleteHistoricalInfo deletes the historical info at a given height
func (k Keeper) DeleteHistoricalInfo(ctx sdk.Context, height int64) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetHistoricalInfoKey(height)

	store.Delete(key)
}
