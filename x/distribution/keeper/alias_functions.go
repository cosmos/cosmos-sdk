package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// get outstanding rewards
func (k Keeper) GetOutstandingRewardsCoins(ctx sdk.Context, val sdk.ValAddress) sdk.DecCoins {
	return k.GetOutstandingRewards(ctx, val)
}

// get the community coins
func (k Keeper) GetFeePoolCommunityCoins(ctx sdk.Context) sdk.DecCoins {
	return k.GetFeePool(ctx).CommunityPool
}
