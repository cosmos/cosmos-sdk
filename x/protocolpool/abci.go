package protocolpool

import (
	"time"

	"cosmossdk.io/x/protocolpool/keeper"
	"cosmossdk.io/x/protocolpool/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	logger := ctx.Logger().With("module", "x/"+types.ModuleName)

	// Iterate over all continuous fund proposals and perform continuous distribution
	k.ContinuousFund.Walk(ctx, nil, func(key sdk.AccAddress, value types.ContinuousFund) (bool, error) {
		cf, err := k.ContinuousFund.Get(ctx, key)
		if err != nil {
			return false, err
		}
		err = k.ContinuousDistribution(ctx, cf)
		if err != nil {
			return false, err
		}

		logger.Info(
			"recipient", key.String(),
			"percentage", cf.Percentage,
		)

		return false, nil
	})

	return nil
}
