package authz

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz/keeper"
)

// MaxExpiredGrantsPerBlock bounds pruning work per block; matches x/feegrant's cap
// on RemoveExpiredAllowances.
const MaxExpiredGrantsPerBlock = 200

// BeginBlocker is called at the beginning of every block
func BeginBlocker(ctx sdk.Context, keeper keeper.Keeper) error {
	return keeper.DequeueAndDeleteExpiredGrants(ctx, MaxExpiredGrantsPerBlock)
}
