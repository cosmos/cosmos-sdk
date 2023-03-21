package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// EndBlocker is called every block, emits balance event
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	k.EmitAllTransientBalances(ctx)
}
