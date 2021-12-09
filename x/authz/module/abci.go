package authz

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz/keeper"
)

// EndBlocker prunes expired authorizations from the state
func EndBlocker(ctx sdk.Context, keeper keeper.Keeper) {

	// delete expired authorizations from store
	keeper.IterateExpiredGrantQueue(ctx, ctx.BlockHeader().Time, func(grantee, granter sdk.AccAddress, msgType string) (stop bool) {
		keeper.DeleteGrant(ctx, grantee, granter, msgType)
		return false
	})
}
