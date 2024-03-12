package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/x/epochs/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
)

// BeginBlocker of epochs module.
func (k Keeper) BeginBlocker(ctx context.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	err := k.IterateEpochInfo(ctx, func(index int64, epochInfo types.EpochInfo) (stop bool) {
		logger := k.Logger()

		// If blocktime < initial epoch start time, return
		if k.environment.HeaderService.GetHeaderInfo(ctx).Time.Before(epochInfo.StartTime) {
			return false
		}
		// if epoch counting hasn't started, signal we need to start.
		shouldInitialEpochStart := !epochInfo.EpochCountingStarted

		epochEndTime := epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
		shouldEpochStart := (k.environment.HeaderService.GetHeaderInfo(ctx).Time.After(epochEndTime)) || shouldInitialEpochStart

		if !shouldEpochStart {
			return false
		}
		epochInfo.CurrentEpochStartHeight = k.environment.HeaderService.GetHeaderInfo(ctx).Height

		if shouldInitialEpochStart {
			epochInfo.EpochCountingStarted = true
			epochInfo.CurrentEpoch = 1
			epochInfo.CurrentEpochStartTime = epochInfo.StartTime
			logger.Info(fmt.Sprintf("Starting new epoch with identifier %s epoch number %d", epochInfo.Identifier, epochInfo.CurrentEpoch))
		} else {
			err := k.environment.EventService.EventManager(ctx).Emit(&types.EventEpochEnd{
				EpochNumber: epochInfo.CurrentEpoch,
			})
			if err != nil {
				return false
			}

			k.AfterEpochEnd(ctx, epochInfo.Identifier, epochInfo.CurrentEpoch)
			epochInfo.CurrentEpoch += 1
			epochInfo.CurrentEpochStartTime = epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
			logger.Info(fmt.Sprintf("Starting epoch with identifier %s epoch number %d", epochInfo.Identifier, epochInfo.CurrentEpoch))
		}

		// emit new epoch start event, set epoch info, and run BeforeEpochStart hook
		err := k.environment.EventService.EventManager(ctx).Emit(&types.EventEpochStart{
			EpochNumber:    epochInfo.CurrentEpoch,
			EpochStartTime: epochInfo.CurrentEpochStartTime.Unix(),
		})
		if err != nil {
			return false
		}
		err = k.setEpochInfo(ctx, epochInfo)
		if err != nil {
			logger.Error(fmt.Sprintf("Error set epoch infor with identifier %s epoch number %d", epochInfo.Identifier, epochInfo.CurrentEpoch))
			return false
		}
		k.BeforeEpochStart(ctx, epochInfo.Identifier, epochInfo.CurrentEpoch)

		return false
	})
	return err
}
