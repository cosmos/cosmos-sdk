package distribution

import (
<<<<<<< HEAD:x/distribution/abci.go
	"time"
=======
	"cosmossdk.io/x/distribution/types"
>>>>>>> 2496cfdf5 (feat: Conditionally emit metrics based on enablement (#19903)):x/distribution/keeper/abci.go

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// BeginBlocker sets the proposer for determining distribution during endblock
// and distribute rewards for the previous block.
<<<<<<< HEAD:x/distribution/abci.go
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
=======
// TODO: use context.Context after including the comet service
func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)
>>>>>>> 2496cfdf5 (feat: Conditionally emit metrics based on enablement (#19903)):x/distribution/keeper/abci.go

	// determine the total power signing the block
	var previousTotalPower int64
	for _, voteInfo := range ctx.VoteInfos() {
		previousTotalPower += voteInfo.Validator.Power
	}

	// TODO this is Tendermint-dependent
	// ref https://github.com/cosmos/cosmos-sdk/issues/3095
	if ctx.BlockHeight() > 1 {
		if err := k.AllocateTokens(ctx, previousTotalPower, ctx.VoteInfos()); err != nil {
			return err
		}
	}

	// record the proposer for when we payout on the next block
	consAddr := sdk.ConsAddress(ctx.BlockHeader().ProposerAddress)
	return k.SetPreviousProposerConsAddr(ctx, consAddr)
}
