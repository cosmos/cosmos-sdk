package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/epochs/types"
)

// GetEpochInfo returns epoch info by identifier.
func (k *Keeper) GetEpochInfo(ctx sdk.Context, identifier string) (types.EpochInfo, error) {
	return k.EpochInfo.Get(ctx, identifier)
}

// AddEpochInfo adds a new epoch info. Will return an error if the epoch fails validation,
// or re-uses an existing identifier.
// This method also sets the start time if left unset, and sets the epoch start height.
func (k *Keeper) AddEpochInfo(ctx sdk.Context, epoch types.EpochInfo) error {
	err := epoch.Validate()
	if err != nil {
		return err
	}
	// Check if identifier already exists
	isExist, err := k.EpochInfo.Has(ctx, epoch.Identifier)
	if err != nil {
		return err
	}
	if isExist {
		return fmt.Errorf("epoch with identifier %s already exists", epoch.Identifier)
	}

	// Initialize empty and default epoch values
	if epoch.StartTime.IsZero() {
		epoch.StartTime = ctx.BlockTime()
	}
	if epoch.CurrentEpochStartHeight == 0 && !epoch.StartTime.After(ctx.BlockTime()) {
		epoch.CurrentEpochStartHeight = ctx.BlockHeight()
	}
	return k.EpochInfo.Set(ctx, epoch.Identifier, epoch)
}

// AllEpochInfos iterate through epochs to return all epochs info.
func (k *Keeper) AllEpochInfos(ctx sdk.Context) ([]types.EpochInfo, error) {
	var epochs []types.EpochInfo
	err := k.EpochInfo.Walk(
		ctx,
		nil,
		func(key string, value types.EpochInfo) (stop bool, err error) {
			epochs = append(epochs, value)
			return false, nil
		},
	)
	return epochs, err
}

// NumBlocksSinceEpochStart returns the number of blocks since the epoch started.
// if the epoch started on block N, then calling this during block N (after BeforeEpochStart)
// would return 0.
// Calling it any point in block N+1 (assuming the epoch doesn't increment) would return 1.
func (k *Keeper) NumBlocksSinceEpochStart(ctx sdk.Context, identifier string) (int64, error) {
	epoch, err := k.EpochInfo.Get(ctx, identifier)
	if err != nil {
		return 0, fmt.Errorf("epoch with identifier %s not found", identifier)
	}
	if ctx.BlockTime().Before(epoch.StartTime) {
		return 0, fmt.Errorf("epoch with identifier %s has not started yet: start time: %s", identifier, epoch.StartTime)
	}

	return ctx.BlockHeight() - epoch.CurrentEpochStartHeight, nil
}
