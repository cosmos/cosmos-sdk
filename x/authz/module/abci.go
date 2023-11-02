package authz

import (
	"context"

	"cosmossdk.io/x/authz/keeper"
)

// BeginBlocker is called at the beginning of every block
func BeginBlocker(ctx context.Context, keeper keeper.Keeper) error {
	// delete all the mature grants
	return keeper.DequeueAndDeleteExpiredGrants(ctx)
}
