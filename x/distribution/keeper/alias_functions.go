package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// get outstanding rewards
func (k Keeper) GetValidatorOutstandingRewardsCoins(ctx sdk.Context, val sdk.ValAddress) sdk.DecCoins {
	return k.GetValidatorOutstandingRewards(ctx, val)
}

// GetFeePoolCommunityCoins get the community pool account coins
func (k Keeper) GetFeePoolCommunityCoins(ctx sdk.Context) sdk.DecCoins {
	communityPool, _ := k.supplyKeeper.GetPoolAccountByName(ctx, CommunityPoolName)
	return communityPool.GetDecCoins()
}
