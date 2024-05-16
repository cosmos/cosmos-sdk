package keeper

import (
	"context"

	"cosmossdk.io/x/distribution/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker sets the proposer for determining distribution during endblock
// and distribute rewards for the previous block.
// TODO: use context.Context after including the comet service
func (k Keeper) BeginBlocker(ctx context.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)
	header := k.HeaderService.HeaderInfo(ctx)
	if header.Height == 0 {
		return nil
	}

	ci := k.cometService.CometInfo(ctx)

	// determine the total power signing the block
	var previousTotalPower int64
	for _, vote := range ci.LastCommit.Votes {
		previousTotalPower += vote.Validator.Power
	}

	// TODO this is Tendermint-dependent
	// ref https://github.com/cosmos/cosmos-sdk/issues/3095
	if header.Height > 1 {
		if err := k.AllocateTokens(ctx, previousTotalPower, ci.LastCommit.Votes); err != nil {
			return err
		}

		// every 1000 blocks send whole coins from decimal pool to community pool
		if header.Height%1000 == 0 {
			if err := k.sendDecimalPoolToCommunityPool(ctx); err != nil {
				return err
			}
		}
	}

	// record the proposer for when we payout on the next block
	consAddr := sdk.ConsAddress(ci.ProposerAddress)
	return k.PreviousProposer.Set(ctx, consAddr)
}
