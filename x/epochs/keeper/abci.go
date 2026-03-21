package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/epochs/types"
)

// BeginBlocker of epochs module.
func (k *Keeper) BeginBlocker(ctx sdk.Context) error {
	start := telemetry.Now()
	defer telemetry.ModuleMeasureSince(types.ModuleName, start, telemetry.MetricKeyBeginBlocker)

	blockTime := ctx.BlockTime()
	blockHeight := ctx.BlockHeight()

	err := k.EpochInfo.Walk(
		ctx,
		nil,
		func(key string, epochInfo types.EpochInfo) (stop bool, err error) {
			// If blocktime < initial epoch start time, return
			if blockTime.Before(epochInfo.StartTime) {
				return false, nil
			}
			// if epoch counting hasn't started, signal we need to start.
			shouldInitialEpochStart := !epochInfo.EpochCountingStarted

			epochEndTime := epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
			shouldEpochStart := (blockTime.After(epochEndTime)) || shouldInitialEpochStart

			if !shouldEpochStart {
				return false, nil
			}
			epochInfo.CurrentEpochStartHeight = blockHeight

			if shouldInitialEpochStart {
				epochInfo.EpochCountingStarted = true
				epochInfo.CurrentEpoch = 1
				epochInfo.CurrentEpochStartTime = epochInfo.StartTime
				ctx.Logger().Debug(fmt.Sprintf("Starting new epoch with identifier %s epoch number %d", epochInfo.Identifier, epochInfo.CurrentEpoch))
			} else {
				err := ctx.EventManager().EmitTypedEvent(&types.EventEpochEnd{
					EpochNumber: epochInfo.CurrentEpoch,
				})
				if err != nil {
					return false, err
				}
				if err != nil {
					return false, nil
				}

				cacheCtx, writeFn := ctx.CacheContext()
				if err := k.AfterEpochEnd(cacheCtx, epochInfo.Identifier, epochInfo.CurrentEpoch); err != nil {
					// purposely ignoring the error here not to halt the chain if the hook fails
					ctx.Logger().Error(fmt.Sprintf("Error after epoch end with identifier %s epoch number %d", epochInfo.Identifier, epochInfo.CurrentEpoch))
				} else {
					writeFn()
				}

				epochInfo.CurrentEpoch += 1
				epochInfo.CurrentEpochStartTime = epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
				ctx.Logger().Debug(fmt.Sprintf("Starting epoch with identifier %s epoch number %d", epochInfo.Identifier, epochInfo.CurrentEpoch))
			}

			// emit new epoch start event, set epoch info, and run BeforeEpochStart hook
			err = ctx.EventManager().EmitTypedEvent(&types.EventEpochStart{
				EpochNumber:    epochInfo.CurrentEpoch,
				EpochStartTime: epochInfo.CurrentEpochStartTime.Unix(),
			})
			if err != nil {
				return false, err
			}
			err = k.EpochInfo.Set(ctx, epochInfo.Identifier, epochInfo)
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("Error set epoch info with identifier %s epoch number %d", epochInfo.Identifier, epochInfo.CurrentEpoch))
				return false, nil
			}

			cacheCtx, writeFn := ctx.CacheContext()
			if err := k.BeforeEpochStart(cacheCtx, epochInfo.Identifier, epochInfo.CurrentEpoch); err != nil {
				// purposely ignoring the error here not to halt the chain if the hook fails
				ctx.Logger().Error(fmt.Sprintf("Error before epoch start with identifier %s epoch number %d", epochInfo.Identifier, epochInfo.CurrentEpoch))
			} else {
				writeFn()
			}

			return false, nil
		},
	)
	return err
}
