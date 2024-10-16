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

func (k Keeper) GetFeePool(ctx context.Context) (feePool types.FeePool) {
	store := k.storeService.OpenKVStore(ctx)
	b, err := store.Get(types.FeePoolKey)
	if b == nil {
		panic("Stored fee pool should not have been nil")
	}
	if err != nil {
		panic(err)
	}
	k.cdc.MustUnmarshal(b, &feePool)

	return
}

// set the global fee pool distribution info
func (k Keeper) SetFeePool(ctx context.Context, feePool types.FeePool) {
	store := k.storeService.OpenKVStore(ctx)
	b := k.cdc.MustMarshal(&feePool)
	store.Set(types.FeePoolKey, b)
}
