package crisis

import (
	"time"

	"cosmossdk.io/x/crisis/keeper"
	"cosmossdk.io/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// check all registered invariants
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	if k.InvCheckPeriod() == 0 || ctx.BlockHeight()%int64(k.InvCheckPeriod()) != 0 {
		// skip running the invariant check
		return
	}
	k.AssertInvariants(ctx)
}
