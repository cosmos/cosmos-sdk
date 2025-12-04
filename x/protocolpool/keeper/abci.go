package keeper

import (
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	start := telemetry.Now()                                                                     //nolint:staticcheck // TODO: switch to OpenTelemetry
	defer telemetry.ModuleMeasureSince(types.ModuleName, start, telemetry.MetricKeyBeginBlocker) //nolint:staticcheck // TODO: switch to OpenTelemetry

	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	if uint64(ctx.BlockHeight())%params.DistributionFrequency == 0 {
		return k.DistributeFunds(ctx)
	}

	return nil
}
