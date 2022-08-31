package distribution

import (
	"strconv"
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
		k.AllocateTokens(ctx, sumPreviousPrecommitPower, previousTotalPower,
			previousProposer, req.LastCommitInfo.GetVotes())

		// validator changes in block n are not reflected in validator power until
		// block n+2, which shows up in the BeginBlocker in block n+3
		// in other words, the validators set that signs n+2 are the validators
		// that had 2/3 power at the end of block n
		// see: https://github.com/tendermint/tendermint/pull/1815
		// e.g.
		// -------------------
		// block n
		// ...
		// -------------------
		// block n+1
		// ...
		// -------------------
		// block n+2
		// block signed from validators finalized in block n
		// ...
		// -------------------
		// block n+3
		// -- begin block --
		// req.LastCommitInfo represents votes from block n+2
		// AllocateTokens pulls BlacklistedPower from block n
		// store blacklisted power for block n+2
		// -- txs --
		// ...

		// Set BlacklistedPower and ValidatorBlacklistedPower
		totalBlacklistedPower, validatorBlacklistedPowers := k.GetValsBlacklistedPowerShare(ctx)
		blacklistedPower := types.BlacklistedPower{
			TotalBlacklistedPowerShare: totalBlacklistedPower,
			ValidatorBlacklistedPowers: validatorBlacklistedPowers,
		}
		// Note: we set the blacklisted power for the previous block height, because
		// the current block hasn't yet processed
		height := strconv.FormatInt(ctx.BlockHeight()-1, 10)
		k.SetBlacklistedPower(ctx, height, blacklistedPower)
	}

	// remove the BlacklistedPower entry for n-4 (which we no longer need)
	k.DeleteBlacklistedPower(ctx, strconv.FormatInt(ctx.BlockHeight()-4, 10))

	// record the proposer for when we payout on the next block
	consAddr := sdk.ConsAddress(req.Header.ProposerAddress)
	k.SetPreviousProposerConsAddr(ctx, consAddr)
}
