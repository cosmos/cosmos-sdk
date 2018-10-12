package distribution

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
)

// set the proposer for determining distribution during endblock
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
	consAddr := sdk.ConsAddress(req.Header.ProposerAddress)
	k.SetProposerConsAddr(ctx, consAddr)

	// determine the total number of signed power
	sumPrecommitPower := int64(0)
	for _, voteInfo := range req.LastCommitInfo.GetVotes() {
		if voteInfo.SignedLastBlock {
			sumPrecommitPower += voteInfo.Validator.Power
		}
	}
	k.SetSumPrecommitPower(ctx, sumPrecommitPower)
}

// allocate fees
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	if ctx.BlockHeight() < 2 {
		return
	}
	k.AllocateFees(ctx)
}
