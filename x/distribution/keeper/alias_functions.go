package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// get outstanding rewards
func (k Keeper) GetValidatorOutstandingRewardsCoins(ctx context.Context, val sdk.ValAddress) (sdk.DecCoins, error) {
	rewards, err := k.GetValidatorOutstandingRewards(ctx, val)
	if err != nil {
		return nil, err
	}

	return rewards.Rewards, nil
}

// get the community coins
func (k Keeper) GetFeePoolCommunityCoins(ctx context.Context) (sdk.DecCoins, error) {
	feePool, err := k.GetFeePool(ctx)
	if err != nil {
		return nil, err
	}

	return feePool.CommunityPool, nil
}

// GetDistributionAccount returns the distribution ModuleAccount
func (k Keeper) GetDistributionAccount(ctx context.Context) sdk.ModuleAccountI {
	return k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
}
