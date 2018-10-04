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
}

// allocate fees
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	if ctx.BlockHeight() < 2 {
		return
	}
	k.AllocateFees(ctx)
}
