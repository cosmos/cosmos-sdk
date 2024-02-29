package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/x/evidence/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker iterates through and handles any newly discovered evidence of
// misbehavior submitted by CometBFT. Currently, only equivocation is handled.
func (k Keeper) BeginBlocker(ctx context.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// If the user is using the legacy CometBFT consensus, we need to convert the
	// evidence to the new types.Misbehavior type and store it in the cache.
	if len(k.evidenceCache) == 0 {
		ci := sdk.UnwrapSDKContext(ctx).CometInfo()
		for _, evidence := range ci.Evidence {
			evi := &types.Misbehavior{
				Type:             types.MisbehaviorType(evidence.Type),
				Height:           evidence.Height,
				Time:             evidence.Time,
				TotalVotingPower: evidence.Validator.Power,
				ConsensusAddress: evidence.Validator.Address,
			}
			k.evidenceCache = append(k.evidenceCache, evi)
		}
	}

	for _, evidence := range k.evidenceCache {
		switch evidence.Type {
		// It's still ongoing discussion how should we treat and slash attacks with
		// premeditation. So for now we agree to treat them in the same way.
		case types.MisbehaviorType_MISBEHAVIOR_TYPE_LIGHT_CLIENT_ATTACK, types.MisbehaviorType_MISBEHAVIOR_TYPE_DUPLICATE_VOTE:
			if evidence == nil {
				continue // skip if no evidence
			}
			evidence := types.FromABCIEvidence(*evidence, k.stakingKeeper.ConsensusAddressCodec())
			err := k.handleEquivocationEvidence(ctx, evidence)
			if err != nil {
				return err
			}
		default:
			k.Logger().Error(fmt.Sprintf("ignored unknown evidence type: %x", evidence.Type))
		}
	}
	return nil
}
