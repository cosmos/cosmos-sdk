package keeper

import (
	"context"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
)

// BeginBlocker will persist the current header and validator set as a historical entry
// and prune the oldest entry based on the HistoricalEntries parameter
func (k *Keeper) BeginBlocker(ctx context.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)
	return k.TrackHistoricalInfo(ctx)
}

// EndBlocker called at every block, update validator set
func (k *Keeper) EndBlocker(ctx context.Context) ([]appmodule.ValidatorUpdate, error) {
	start := telemetry.Now()
	defer telemetry.ModuleMeasureSince(types.ModuleName, start, telemetry.MetricKeyEndBlocker)
	return k.BlockValidatorUpdates(ctx)
}
