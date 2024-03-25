package crisis

import (
	"context"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	"github.com/cosmos/cosmos-sdk/x/crisis/types"
)

// check all registered invariants
func EndBlocker(ctx context.Context, k keeper.Keeper) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer telemetry.ModuleMeasureSince(sdkCtx, types.ModuleName, telemetry.Now(sdkCtx), telemetry.MetricKeyEndBlocker)

	if k.InvCheckPeriod() == 0 || sdkCtx.BlockHeight()%int64(k.InvCheckPeriod()) != 0 {
		// skip running the invariant check
		return
	}
	k.AssertInvariants(sdkCtx)
}
