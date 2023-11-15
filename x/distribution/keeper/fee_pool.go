package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// DistributeFromFeePool distributes funds from the distribution module account to
// a receiver address while updating the community pool
func (k Keeper) DistributeFromFeePool(ctx context.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error {
	feePool, err := k.FeePool.Get(ctx)
	if err != nil {
		return err
	}

	// NOTE the community pool isn't a module account, however its coins
	// are held in the distribution module account. Thus the community pool
	// must be reduced separately from the SendCoinsFromModuleToAccount call
	newPool, negative := feePool.CommunityPool.SafeSub(sdk.NewDecCoinsFromCoins(amount...))
	if negative {
		return types.ErrBadDistribution
	}

	feePool.CommunityPool = newPool

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiveAddr, amount)
	if err != nil {
		return err
	}

	return k.FeePool.Set(ctx, feePool)
}

// DistributeLiquidityProviderReward sends funds from the liquidity provider reward pool to
// a receiver address, and updates the remaining amount in the pool.
func (k Keeper) DistributeLiquidityProviderReward(ctx context.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error {
	feePool, err := k.FeePool.Get(ctx)
	if err != nil {
		return err
	}

	// NOTE the lp reward pool isn't a module account, however its coins
	// are held in the distribution module account. Thus the community pool
	// must be reduced separately from the SendCoinsFromModuleToAccount call
	newPool, negative := feePool.LiquidityProviderPool.SafeSub(sdk.NewDecCoinsFromCoins(amount...))
	if negative {
		return types.ErrBadDistribution
	}

	feePool.LiquidityProviderPool = newPool

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiveAddr, amount)
	if err != nil {
		return err
	}

	err = k.FeePool.Set(ctx, feePool)
	if err != nil {
		return err
	}
	return nil
}
