package distribution

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
)

// set the proposer for determining distribution during endblock
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {

	if ctx.BlockHeight() > 1 {
		previousPercentPrecommitVotes := getPreviousPercentPrecommitVotes(req)
		previousProposer := k.GetPreviousProposerConsAddr(ctx)
		k.AllocateTokens(ctx, previousPercentPrecommitVotes, previousProposer)
	}

	consAddr := sdk.ConsAddress(req.Header.ProposerAddress)
	k.SetPreviousProposerConsAddr(ctx, consAddr)
}

// set the total power and validator power's for next block
func EndBlocker(ctx sdk.Context, k keeper.Keeper, validatorUpdates []sdk.ValidatorUpdate, totalPower int64) {
	k.SetPreviousTotalPower(ctx, totalPower)

	for _, vu := range validatorUpdates {
		k.SetPreviousValidatorPower(ctx, vu.Address, vu.Power)
	}
}

// percent precommit votes for the previous block
func getPreviousPercentPrecommitVotes(req abci.RequestBeginBlock) sdk.Dec {

	// determine the total number of signed power
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
