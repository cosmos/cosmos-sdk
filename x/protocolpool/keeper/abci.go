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

	logger := k.Logger(ctx)

	// Iterate over all continuous fund proposals and perform continuous distribution
	// Note:  This loop intentionally processes all payment streams, and the number of streams
	// can impact block processing time. However, since it is governed by governance, it is not
	// considered a denial-of-service (DoS) factor.
	err := k.ContinuousFund.Walk(ctx, nil, func(key sdk.AccAddress, value types.ContinuousFund) (bool, error) {
		err := k.continuousDistribution(ctx, value)
		if err != nil {
			return false, err
		}

		logger.Debug(
			"recipient", key.String(),
			"percentage", value.Percentage,
		)

		return false, nil
	})
	if err != nil {
		return err
	}

	return nil
}
