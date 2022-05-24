package crisis

import (
	"time"

	"github.com/Stride-Labs/cosmos-sdk/telemetry"
	sdk "github.com/Stride-Labs/cosmos-sdk/types"
	"github.com/Stride-Labs/cosmos-sdk/x/crisis/keeper"
	"github.com/Stride-Labs/cosmos-sdk/x/crisis/types"
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
