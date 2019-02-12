package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// get outstanding rewards
func (k Keeper) GetOutstandingRewardsCoins(ctx sdk.Context) sdk.DecCoins {
	return GetOutstandingRewards
}

// get the community coins
func (k Keeper) GetOutstandingRewardsCoins(ctx sdk.Context) sdk.DecCoins {
	return GetFeePool.CommunityPool
}
