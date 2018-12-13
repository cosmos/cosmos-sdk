package distribution

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
)

// set the proposer for determining distribution during endblock
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {

	// TODO This is Tendermint-dependent. ref https://github.com/cosmos/cosmos-sdk/issues/3095
	if ctx.BlockHeight() > 1 {
		previousFractionPrecommitVotes := getFractionPercentPrecommitVotes(req)
		previousProposer := k.GetPreviousProposerConsAddr(ctx)
		k.AllocateTokens(ctx, previousFractionPrecommitVotes, previousProposer)
	}

	consAddr := sdk.ConsAddress(req.Header.ProposerAddress)
	k.SetPreviousProposerConsAddr(ctx, consAddr)
}

// fraction precommit votes for the previous block
func getFractionPercentPrecommitVotes(req abci.RequestBeginBlock) sdk.Dec {

	// determine the total power signing the block
	totalPower, sumPrecommitPower := int64(0), int64(0)
	for _, voteInfo := range req.LastCommitInfo.GetVotes() {
		totalPower += voteInfo.Validator.Power
		if voteInfo.SignedLastBlock {
			sumPrecommitPower += voteInfo.Validator.Power
		}
	}

	if totalPower == 0 {
		return sdk.ZeroDec()
	}

	return sdk.NewDec(sumPrecommitPower).Quo(sdk.NewDec(totalPower))
}
