package crisis

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/crisis/keeper"
)

// check all registered invariants
func EndBlocker(ctx context.Context, k keeper.Keeper) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if k.InvCheckPeriod() == 0 || sdkCtx.BlockHeight()%int64(k.InvCheckPeriod()) != 0 {
		// skip running the invariant check
		return
	}
	k.AssertInvariants(sdkCtx)
}
