package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/core/info"
	"cosmossdk.io/x/evidence/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker iterates through and handles any newly discovered evidence of
// misbehavior submitted by CometBFT. Currently, only equivocation is handled.
func (k Keeper) BeginBlocker(goCtx context.Context) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	bi := k.cometInfo.GetCometInfo(goCtx)

	ctx := sdk.UnwrapSDKContext(goCtx)
	for _, tmEvidence := range bi.GetMisbehavior() {
		switch tmEvidence.Type {
		// It's still ongoing discussion how should we treat and slash attacks with
		// premeditation. So for now we agree to treat them in the same way.
		case info.LightClientAttack, info.DuplicateVote:
			evidence := types.FromABCIEvidence(tmEvidence)
			k.handleEquivocationEvidence(ctx, evidence)

		default:
			k.Logger(ctx).Error(fmt.Sprintf("ignored unknown evidence type: %x", tmEvidence.Type))
		}
	}
}
