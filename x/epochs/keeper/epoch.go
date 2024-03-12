package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/x/epochs/types"
)

// GetEpochInfo returns epoch info by identifier.
func (k Keeper) GetEpochInfo(ctx context.Context, identifier string) types.EpochInfo {
	epoch, err := k.EpochInfo.Get(ctx, identifier)
	if err != nil {
		return types.EpochInfo{}
	}
	return epoch
}

// AddEpochInfo adds a new epoch info. Will return an error if the epoch fails validation,
// or re-uses an existing identifier.
// This method also sets the start time if left unset, and sets the epoch start height.
func (k Keeper) AddEpochInfo(ctx context.Context, epoch types.EpochInfo) error {
	err := epoch.Validate()
	if err != nil {
		return err
	}
	// Check if identifier already exists
	if (k.GetEpochInfo(ctx, epoch.Identifier) != types.EpochInfo{}) {
		return fmt.Errorf("epoch with identifier %s already exists", epoch.Identifier)
	}

	// Initialize empty and default epoch values
	if epoch.StartTime.Equal(time.Time{}) {
		epoch.StartTime = k.environment.HeaderService.GetHeaderInfo(ctx).Time
	}
	epoch.CurrentEpochStartHeight = k.environment.HeaderService.GetHeaderInfo(ctx).Height

	err = k.setEpochInfo(ctx, epoch)
	return err
}

// setEpochInfo set epoch info.
func (k Keeper) setEpochInfo(ctx context.Context, epoch types.EpochInfo) error {
	err := k.EpochInfo.Set(ctx, epoch.Identifier, epoch)
	return err
}

// DeleteEpochInfo delete epoch info.
func (k Keeper) DeleteEpochInfo(ctx context.Context, identifier string) error {
	err := k.EpochInfo.Remove(ctx, identifier)
	return err
}

// IterateEpochInfo iterate through epochs.
func (k Keeper) IterateEpochInfo(ctx context.Context, fn func(index int64, epochInfo types.EpochInfo) (stop bool)) error {
	iter, err := k.EpochInfo.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	i := int64(0)

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}

		stop := fn(i, kv.Value)

		if stop {
			break
		}
		i++
	}
	return nil
}

// AllEpochInfos iterate through epochs to return all epochs info.
func (k Keeper) AllEpochInfos(ctx context.Context) ([]types.EpochInfo, error) {
	epochs := []types.EpochInfo{}
	err := k.IterateEpochInfo(ctx, func(index int64, epochInfo types.EpochInfo) (stop bool) {
		epochs = append(epochs, epochInfo)
		return false
	})
	return epochs, err
}

// NumBlocksSinceEpochStart returns the number of blocks since the epoch started.
// if the epoch started on block N, then calling this during block N (after BeforeEpochStart)
// would return 0.
// Calling it any point in block N+1 (assuming the epoch doesn't increment) would return 1.
func (k Keeper) NumBlocksSinceEpochStart(ctx context.Context, identifier string) (int64, error) {
	epoch := k.GetEpochInfo(ctx, identifier)
	if (epoch == types.EpochInfo{}) {
		return 0, fmt.Errorf("epoch with identifier %s not found", identifier)
	}
	return k.environment.HeaderService.GetHeaderInfo(ctx).Height - epoch.CurrentEpochStartHeight, nil
}
