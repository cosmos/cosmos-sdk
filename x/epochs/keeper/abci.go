package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/x/epochs/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
)

// BeginBlocker of epochs module.
func (k Keeper) BeginBlocker(ctx context.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	logger := k.Logger()
	headerInfo := k.environment.HeaderService.GetHeaderInfo(ctx)
	err := k.EpochInfo.Walk(
		ctx,
		nil,
		func(key string, epochInfo types.EpochInfo) (stop bool, err error) {
			// If blocktime < initial epoch start time, return
			if headerInfo.Time.Before(epochInfo.StartTime) {
				return false, nil
			}
			// if epoch counting hasn't started, signal we need to start.
			shouldInitialEpochStart := !epochInfo.EpochCountingStarted

			epochEndTime := epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
			shouldEpochStart := (headerInfo.Time.After(epochEndTime)) || shouldInitialEpochStart

			if !shouldEpochStart {
				return false, nil
			}
			epochInfo.CurrentEpochStartHeight = headerInfo.Height

			if shouldInitialEpochStart {
				epochInfo.EpochCountingStarted = true
				epochInfo.CurrentEpoch = 1
				epochInfo.CurrentEpochStartTime = epochInfo.StartTime
				logger.Debug(fmt.Sprintf("Starting new epoch with identifier %s epoch number %d", epochInfo.Identifier, epochInfo.CurrentEpoch))
			} else {
				err := k.environment.EventService.EventManager(ctx).Emit(&types.EventEpochEnd{
					EpochNumber: epochInfo.CurrentEpoch,
				})
				if err != nil {
					return false, nil
				}

				if err := k.environment.BranchService.Execute(ctx, func(ctx context.Context) error {
					return k.AfterEpochEnd(ctx, epochInfo.Identifier, epochInfo.CurrentEpoch)
				}); err != nil {
					// purposely ignoring the error here not to halt the chain if the hook fails
					logger.Error(fmt.Sprintf("Error after epoch end with identifier %s epoch number %d", epochInfo.Identifier, epochInfo.CurrentEpoch))
				}

				epochInfo.CurrentEpoch += 1
				epochInfo.CurrentEpochStartTime = epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
				logger.Debug(fmt.Sprintf("Starting epoch with identifier %s epoch number %d", epochInfo.Identifier, epochInfo.CurrentEpoch))
			}

			// emit new epoch start event, set epoch info, and run BeforeEpochStart hook
			err = k.environment.EventService.EventManager(ctx).Emit(&types.EventEpochStart{
				EpochNumber:    epochInfo.CurrentEpoch,
				EpochStartTime: epochInfo.CurrentEpochStartTime.Unix(),
			})
			if err != nil {
				return false, err
			}
			err = k.EpochInfo.Set(ctx, epochInfo.Identifier, epochInfo)
			if err != nil {
				logger.Error(fmt.Sprintf("Error set epoch info with identifier %s epoch number %d", epochInfo.Identifier, epochInfo.CurrentEpoch))
				return false, nil
			}
			if err := k.environment.BranchService.Execute(ctx, func(ctx context.Context) error {
				return k.BeforeEpochStart(ctx, epochInfo.Identifier, epochInfo.CurrentEpoch)
			}); err != nil {
				// purposely ignoring the error here not to halt the chain if the hook fails
				logger.Error(fmt.Sprintf("Error before epoch start with identifier %s epoch number %d", epochInfo.Identifier, epochInfo.CurrentEpoch))
			}
			return false, nil
		},
	)
	return err
}
