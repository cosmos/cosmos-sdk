package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// BeginBlocker sets the proposer for determining distribution during endblock
// and distribute rewards for the previous block.
func (k Keeper) BeginBlocker(ctx context.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)
	c := sdk.UnwrapSDKContext(ctx)

	// dynamically adjust the nakamoto bonus coefficient first.
	if err := k.AdjustNakamotoBonusCoefficient(c); err != nil {
		return err
	}

	// determine the total power signing the block
	var previousTotalPower int64
	for _, voteInfo := range c.VoteInfos() {
		previousTotalPower += voteInfo.Validator.Power
	}

	// TODO this is Tendermint-dependent
	// ref https://github.com/cosmos/cosmos-sdk/issues/3095
	if c.BlockHeight() > 1 {
		if err := k.AllocateTokens(ctx, previousTotalPower, c.VoteInfos()); err != nil {
			return err
		}
	}

	// record the proposer for when we payout on the next block
	consAddr := sdk.ConsAddress(c.BlockHeader().ProposerAddress)
	return k.SetPreviousProposerConsAddr(ctx, consAddr)
}
