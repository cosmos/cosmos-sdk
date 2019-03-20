package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// get outstanding rewards
func (k Keeper) GetValidatorOutstandingRewardsCoins(ctx sdk.Context, val sdk.ValAddress) sdk.DecCoins {
	return k.GetValidatorOutstandingRewards(ctx, val)
}

// get the community coins
func (k Keeper) GetFeePoolCommunityCoins(ctx sdk.Context) sdk.DecCoins {
	return k.GetFeePool(ctx).CommunityPool
}
