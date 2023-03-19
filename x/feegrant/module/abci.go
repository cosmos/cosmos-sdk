package module

import (
	"cosmossdk.io/x/feegrant/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	k.RemoveExpiredAllowances(ctx)
}
