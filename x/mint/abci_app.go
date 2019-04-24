package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/keeper"
)

// BeginBlocker inflates every block and updates inflation parameters once per hour
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {

	k.CalculateInflationRate(ctx)
	k.Mint(ctx)
}
