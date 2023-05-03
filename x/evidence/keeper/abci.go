package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/core/comet"
	"cosmossdk.io/x/evidence/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker iterates through and handles any newly discovered evidence of
// misbehavior submitted by CometBFT. Currently, only equivocation is handled.
func (k Keeper) BeginBlocker(goCtx context.Context) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	bi := k.cometInfo.GetCometBlockInfo(goCtx).GetEvidence()

	for i := 0; i < bi.Len(); i++ {
		switch bi.Get(i).Type() {
		// It's still ongoing discussion how should we treat and slash attacks with
		// premeditation. So for now we agree to treat them in the same way.
		case comet.LightClientAttack, comet.DuplicateVote:
			evidence := types.FromABCIEvidence(bi.Get(i))
			k.handleEquivocationEvidence(goCtx, evidence)

		default:
			ctx := sdk.UnwrapSDKContext(goCtx)
			k.Logger(ctx).Error(fmt.Sprintf("ignored unknown evidence type: %x", bi.Get(i).Type()))
		}
	}
}
