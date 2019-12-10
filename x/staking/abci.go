package staking

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// BeginBlocker will
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
	if ctx.BlockHeight() > int64(entryNum) {
		k.DeleteHistoricalInfo(ctx, ctx.BlockHeight()-int64(entryNum))
	}
}
