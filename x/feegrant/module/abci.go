package module

import (
	"context"

	"cosmossdk.io/x/feegrant/keeper"
)

func EndBlocker(ctx context.Context, k keeper.Keeper) error {
	return k.RemoveExpiredAllowances(ctx, 200)
}
