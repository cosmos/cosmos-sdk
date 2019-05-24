package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// get outstanding rewards
func (k Keeper) GetValidatorOutstandingRewardsCoins(ctx sdk.Context, val sdk.ValAddress) sdk.DecCoins {
	return k.GetValidatorOutstandingRewards(ctx, val)
}

// get the community coins
func (k Keeper) GetFeePoolCommunityCoins(ctx sdk.Context) sdk.DecCoins {
	return k.GetFeePool(ctx).CommunityPool
}

// GetPoolAccountByName alias for supply keeper's GetPoolAccountByName
func (k Keeper) GetPoolAccountByName(ctx sdk.Context, name string) supply.PoolAccount {
	return k.supplyKeeper.GetPoolAccountByName(ctx, name)
}

// SetPoolAccount alias for supply keeper's SetPoolAccount
func (k Keeper) SetPoolAccount(ctx sdk.Context, pAcc supply.PoolAccount) {
	k.supplyKeeper.SetPoolAccount(ctx, pAcc)
}
