package authz

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz/keeper"
)

// EndBlocker is called at the end of every block
func EndBlocker(ctx sdk.Context, keeper keeper.Keeper) {

	// clears all the mature grants
	keeper.DeleteAllMatureGrants(ctx)
}
