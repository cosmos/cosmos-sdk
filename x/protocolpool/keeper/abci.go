package keeper

import (
	"context"
	"time"

	"cosmossdk.io/x/protocolpool/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k *Keeper) EndBlocker(ctx context.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	logger := k.Logger(ctx).With("module", "x/"+types.ModuleName)

	// Iterate over all continuous fund proposals and perform continuous distribution
	err := k.ContinuousFund.Walk(ctx, nil, func(key sdk.AccAddress, value types.ContinuousFund) (bool, error) {
		cf, err := k.ContinuousFund.Get(ctx, key)
		if err != nil {
			return false, err
		}
		err = k.continuousDistribution(ctx, cf)
		if err != nil {
			return false, err
		}

		logger.Info(
			"recipient", key.String(),
			"percentage", cf.Percentage,
		)

		return false, nil
	})
	if err != nil {
		return err
	}

	return nil
}
