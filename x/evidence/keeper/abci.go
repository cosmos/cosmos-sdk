package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/core/blockinfo"
	"cosmossdk.io/x/evidence/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker iterates through and handles any newly discovered evidence of
// misbehavior submitted by CometBFT. Currently, only equivocation is handled.
func (k Keeper) BeginBlocker(goctx context.Context) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	ctx := sdk.UnwrapSDKContext(goctx)
	for _, tmEvidence := range k.blockInfo.Misbehavior() {
		switch tmEvidence.Type {
		// It's still ongoing discussion how should we treat and slash attacks with
		// premeditation. So for now we agree to treat them in the same way.
		case blockinfo.LightClientAttack, blockinfo.DuplicateVote:
			evidence := types.FromABCIEvidence(tmEvidence)
			k.handleEquivocationEvidence(ctx, evidence)

		default:
			k.Logger(ctx).Error(fmt.Sprintf("ignored unknown evidence type: %x", tmEvidence.Type))
		}
	}
}
