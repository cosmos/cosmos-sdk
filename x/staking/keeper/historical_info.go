package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TrackHistoricalInfo saves the latest historical-info and deletes the oldest
// heights that are below pruning height
func (k Keeper) TrackHistoricalInfo(ctx context.Context) error {
	entryNum, err := k.HistoricalEntries(ctx)
	if err != nil {
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Prune store to ensure we only have parameter-defined historical entries.
	// In most cases, this will involve removing a single historical entry.
	// In the rare scenario when the historical entries gets reduced to a lower value k'
	// from the original value k. k - k' entries must be deleted from the store.
	// Since the entries to be deleted are always in a continuous range, we can iterate
	// over the historical entries starting from the most recent version to be pruned
	// and then return at the first empty entry.
	for i := sdkCtx.HeaderInfo().Height - int64(entryNum); i >= 0; i-- {
		has, err := k.HistoricalInfo.Has(ctx, uint64(i))
		if err != nil {
			return err
		}
		if !has {
			break
		}
		if err = k.HistoricalInfo.Remove(ctx, uint64(i)); err != nil {
			return err
		}
	}

	// if there is no need to persist historicalInfo, return
	if entryNum == 0 {
		return nil
	}

	time := sdkCtx.HeaderInfo().Time

	historicalEntry := types.HistoricalRecord{
		Time:           &time,
		ValidatorsHash: sdkCtx.CometInfo().ValidatorsHash,
		Apphash:        sdkCtx.HeaderInfo().AppHash,
	}

	// Set latest HistoricalInfo at current height
	return k.HistoricalInfo.Set(ctx, uint64(sdkCtx.HeaderInfo().Height), historicalEntry)
}
