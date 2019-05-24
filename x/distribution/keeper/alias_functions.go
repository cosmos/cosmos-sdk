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

// GetModuleAccountByName alias for supply keeper's GetModuleAccountByName
func (k Keeper) GetModuleAccountByName(ctx sdk.Context, name string) supply.ModuleAccount {
	return k.supplyKeeper.GetModuleAccountByName(ctx, name)
}

// SetModuleAccount alias for supply keeper's SetModuleAccount
func (k Keeper) SetModuleAccount(ctx sdk.Context, pAcc supply.ModuleAccount) {
	k.supplyKeeper.SetModuleAccount(ctx, pAcc)
}
