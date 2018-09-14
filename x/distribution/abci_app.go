package distribution

import (
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	abci "github.com/tendermint/tendermint/abci/types"
)

// set the proposer for determining distribution during endblock
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
	consAddr := sdk.ConsAddress{req.Header.Proposer.Address}
	k.SetProposerConsAddr(consAddr)
}

// allocate fees
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	k.AllocateFees(ctx)
}
