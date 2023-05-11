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

	bi := k.cometInfo.GetCometBlockInfo(goCtx)
	if bi == nil {
		// If we don't have block info, we don't have any evidence to process.  Block info may be nil during
		// genesis calls or in tests.
		return
	}

	evidences := bi.GetEvidence()

	ctx := sdk.UnwrapSDKContext(goCtx)
	for i := 0; i < evidences.Len(); i++ {
		switch evidences.Get(i).Type() {
		// It's still ongoing discussion how should we treat and slash attacks with
		// premeditation. So for now we agree to treat them in the same way.
		case comet.LightClientAttack, comet.DuplicateVote:
			evidence := types.FromABCIEvidence(evidences.Get(i))
			k.handleEquivocationEvidence(ctx, evidence)

		default:
			k.Logger(ctx).Error(fmt.Sprintf("ignored unknown evidence type: %x", evidences.Get(i).Type()))
		}
	}
}
