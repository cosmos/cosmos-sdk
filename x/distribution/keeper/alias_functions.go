package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
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

// GetDistributionAccount returns the distribution ModuleAccount
func (k Keeper) GetDistributionAccount(ctx sdk.Context) supply.ModuleAccount {
	return k.supplyKeeper.GetModuleAccountByName(ctx, types.ModuleName)
}

// SetDistributionAccount alias for supply keeper's SetModuleAccount
func (k Keeper) SetDistributionAccount(ctx sdk.Context, macc supply.ModuleAccount) {
	if macc.Name() != types.ModuleName {
		panic("cannot set a module account other than distribution's")
	}
	k.supplyKeeper.SetModuleAccount(ctx, macc)
}
