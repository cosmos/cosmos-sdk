package tieredfee

import (
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/tieredfee/keeper"
	"github.com/cosmos/cosmos-sdk/x/tieredfee/types"
)

// BeginBlocker will update gas prices for all tiers according to block gas used recorded in last end blocker.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	k.UpdateAllTiers(ctx)
}

// EndBlocker will update block gas used.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	k.SetBlockGasUsed(ctx, ctx.BlockGasMeter().GasConsumed())
	return []abci.ValidatorUpdate{}
}
