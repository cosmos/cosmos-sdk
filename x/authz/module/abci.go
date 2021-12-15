package authz

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz/keeper"
)

// EndBlocker is called at the end of every block
func EndBlocker(ctx sdk.Context, keeper keeper.Keeper) {

	matureGrants, err := keeper.DequeueAllMatureGrants(ctx)
	if err != nil {
		panic(err)
	}

	// clears all the mature grants
	if err := keeper.DeleteExpiredGrants(ctx, matureGrants); err != nil {
		panic(err)
	}
}
