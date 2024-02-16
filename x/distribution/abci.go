package distribution

import (
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

var BlockMultipleToDistributeRewards = int64(50)

// BeginBlocker sets the proposer for determining distribution during endblock
// and distribute rewards for the previous block.
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// TODO this is Tendermint-dependent
	// ref https://github.com/cosmos/cosmos-sdk/issues/3095
	blockHeight := ctx.BlockHeight()
	// only allocate rewards if the block height is greater than 1
	// and for every multiple of 50 blocks for performance reasons.
	if blockHeight > 1 && blockHeight%BlockMultipleToDistributeRewards == 0 {
		// determine the total power signing the block
		var previousTotalPower int64
		for _, voteInfo := range req.LastCommitInfo.GetVotes() {
			previousTotalPower += voteInfo.Validator.Power
		}

		k.AllocateTokens(ctx, previousTotalPower, req.LastCommitInfo.GetVotes())
	}

	// record the proposer for when we payout on the next block
	consAddr := sdk.ConsAddress(req.Header.ProposerAddress)
	k.SetPreviousProposerConsAddr(ctx, consAddr)
}
