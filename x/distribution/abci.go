package distribution

import (
	"fmt"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// BeginBlocker sets the proposer for determining distribution during endblock
// and distribute rewards for the previous block
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// determine the total power signing the block
	var previousTotalPower, sumPreviousPrecommitPower int64
	for _, voteInfo := range req.LastCommitInfo.GetVotes() {
		previousTotalPower += voteInfo.Validator.Power
		if voteInfo.SignedLastBlock {
			sumPreviousPrecommitPower += voteInfo.Validator.Power
		}
	}

	// TODO this is Tendermint-dependent
	// ref https://github.com/cosmos/cosmos-sdk/issues/3095
	if ctx.BlockHeight() > 1 {
		previousProposer := k.GetPreviousProposerConsAddr(ctx)
		k.AllocateTokens(ctx, sumPreviousPrecommitPower, previousTotalPower, previousProposer, req.LastCommitInfo.GetVotes())
	}

	restakeFunc := func(delegator sdk.AccAddress, validator sdk.ValAddress) (stop bool) {
		err := k.PerformRestake(ctx, delegator, validator)
		if err != nil {
			k.Logger(ctx).Info(fmt.Sprintf("Err: %s, Failed to perform restake for delegator-validator %s - %s", err, delegator, validator))
		}

		return err != nil
	}

	if ctx.BlockHeight()%k.GetRestakePeriod(ctx).Int64() == 0 {
		staleKeys := k.IterateRestakeEntries(ctx, restakeFunc)

		for _, stale := range staleKeys {

			err := k.DeleteAutoRestakeEntry(ctx, stale.Delegator, stale.Validator)
			if err != nil {
				k.Logger(ctx).Info(fmt.Sprintf("Err: %s, Failed to delete restake key", err))
			}
		}
	}

	// record the proposer for when we payout on the next block
	consAddr := sdk.ConsAddress(req.Header.ProposerAddress)
	k.SetPreviousProposerConsAddr(ctx, consAddr)
}
