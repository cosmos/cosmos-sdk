package keeper

import (
	"context"
	"errors"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// IterateHistoricalInfo provides an iterator over all stored HistoricalInfo
// objects. For each HistoricalInfo object, cb will be called. If the cb returns
// true, the iterator will break and close.
func (k Keeper) IterateHistoricalInfo(ctx context.Context, cb func(types.HistoricalInfo) bool) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.HistoricalInfoKey, storetypes.PrefixEndBytes(types.HistoricalInfoKey))
	if err != nil {
		return err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		histInfo, err := types.UnmarshalHistoricalInfo(k.cdc, iterator.Value())
		if err != nil {
			return err
		}
		if cb(histInfo) {
			break
		}
	}

	return nil
}

// GetAllHistoricalInfo returns all stored HistoricalInfo objects.
func (k Keeper) GetAllHistoricalInfo(ctx context.Context) ([]types.HistoricalInfo, error) {
	var infos []types.HistoricalInfo
	err := k.HistoricalInfo.Walk(ctx, nil, func(key []byte, info types.HistoricalInfo) (stop bool, err error) {
		infos = append(infos, info)
		return false, nil
	})
	if err != nil {
		return []types.HistoricalInfo{}, err
	}

	return infos, nil
}

// GetHistoricalInfo gets the historical info at a given height
func (k Keeper) GetHistoricalInfo(ctx context.Context, height int64) (types.HistoricalInfo, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetHistoricalInfoKey(height)

	value, err := store.Get(key)
	if err != nil {
		return types.HistoricalInfo{}, err
	}

	if value == nil {
		return types.HistoricalInfo{}, types.ErrNoHistoricalInfo
	}

	return types.UnmarshalHistoricalInfo(k.cdc, value)
}

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
	for i := sdkCtx.BlockHeight() - int64(entryNum); i >= 0; i-- {
		_, err := k.GetHistoricalInfo(ctx, i)
		if err != nil {
			if errors.Is(err, types.ErrNoHistoricalInfo) {
				break
			}
			return err
		}
		if err = k.HistoricalInfo.Remove(ctx, types.GetHistoricalInfoKeyWithoutPrefix(i)); err != nil {
			return err
		}
	}

	// if there is no need to persist historicalInfo, return
	if entryNum == 0 {
		return nil
	}

	// Create HistoricalInfo struct
	lastVals, err := k.GetLastValidators(ctx)
	if err != nil {
		return err
	}

	historicalEntry := types.NewHistoricalInfo(sdkCtx.BlockHeader(), lastVals, k.PowerReduction(ctx))
	err = k.HistoricalInfo.Set(ctx, types.GetHistoricalInfoKeyWithoutPrefix(sdkCtx.BlockHeight()), historicalEntry)
	// Set latest HistoricalInfo at current height
	return err
}
