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
		//   block (n)	  			begin blocker (n+1) 										  block (n+1)         begin blocker (n+2)
		//   ---- n ---- | --- n + 1 --- distribute(tokens, bp_n, vpb_n), set(bp_n+1 + vpb_n+1) | --- txs ---- | --- n + 2 --- distribute(tokens, bp, vpb), set(bp + vpb)
		// Set BlacklistedPower and ValidatorBlacklistedPower
		totalBlacklistedPower, validatorBlacklistedPowers := k.GetValsBlacklistedPowerShare(ctx)
		blacklistedPower := types.BlacklistedPower{
			TotalBlacklistedPowerShare: totalBlacklistedPower,
			ValidatorBlacklistedPowers: validatorBlacklistedPowers,
		}
		height := strconv.FormatInt(ctx.BlockHeight(), 10)
		k.SetBlacklistedPower(ctx, height, blacklistedPower)
	}

	// record the proposer for when we payout on the next block
	consAddr := sdk.ConsAddress(req.Header.ProposerAddress)
	k.SetPreviousProposerConsAddr(ctx, consAddr)
}
