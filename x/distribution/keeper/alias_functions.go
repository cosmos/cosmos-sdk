package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// get outstanding rewards
func (k Keeper) GetValidatorOutstandingRewardsCoins(ctx sdk.Context, val sdk.ValAddress) sdk.DecCoins {
	return k.GetValidatorOutstandingRewards(ctx, val)
}

// GetFeePoolCommunityCoins get the community pool account coins
func (k Keeper) GetFeePoolCommunityCoins(ctx sdk.Context) sdk.Coins {
	communityPool, err := k.supplyKeeper.GetPoolAccountByName(ctx, CommunityPoolName)
	if err != nil {
		panic(err)
	}
	return communityPool.GetCoins()
}
