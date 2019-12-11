package staking

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// BeginBlocker will persist the current header and validator set as a historical entry
// and prune the oldest entry based on the HistoricalEntries parameter
func BeginBlocker(ctx sdk.Context, k Keeper) {
	entryNum := k.HistoricalEntries(ctx)
	// if there is no need to persist historicalInfo, return
	if entryNum == 0 {
		return
	}

	// Create HistoricalInfo struct
	lastVals := k.GetLastValidators(ctx)
	types.Validators(lastVals).Sort()
	historicalEntry := types.HistoricalInfo{
		Header: ctx.BlockHeader(),
		ValSet: lastVals,
	}

	// Set latest HistoricalInfo at current height
	k.SetHistoricalInfo(ctx, ctx.BlockHeight(), historicalEntry)

	// prune store to ensure we only have parameter-defined historical entries
	// in most cases, this will invole removing a single historical entry
	// In the rare scenario when the historical entries gets reduced to a lower value k'
	// from the original value k. k - k' entries must be deleted from the store
	// Since the entries to be deleted are always in a continuous range, we can iterate
	// over the historical entries starting from the most recent version to be pruned
	// and then return at the first empty entry
	for i := ctx.BlockHeight() - int64(entryNum); i >= 0; i-- {
		_, found := k.GetHistoricalInfo(ctx, i)
		if found {
			k.DeleteHistoricalInfo(ctx, i)
		} else {
			return
		}
	}
}
