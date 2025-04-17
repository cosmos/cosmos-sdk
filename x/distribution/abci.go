package distribution

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
)

// BeginBlocker sets the proposer for determining distribution during endblock
// and distribute rewards for the previous block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) error {
	if err := k.BeginBlocker(ctx); err != nil {
		ctx.Logger().Error("BeginBlocker failed", "error", err)
		return err
	}

	return nil
}
