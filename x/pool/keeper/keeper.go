package keeper

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/x/pool/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	storeService storetypes.KVStoreService
	authKeeper   types.AccountKeeper
	bankKeeper   types.BankKeeper

	authority string
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService,
	ak types.AccountKeeper, bk types.BankKeeper, authority string,
) Keeper {
	// ensure pool module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}
	return Keeper{
		storeService: storeService,
		authKeeper:   ak,
		bankKeeper:   bk,
		authority:    authority,
	}
}

// GetAuthority returns the x/pool module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With(log.ModuleKey, "x/"+types.ModuleName)
}

// FundCommunityPool allows an account to directly fund the community fund pool.
// The amount is first added to the distribution module account and then directly
// added to the pool. An error is returned if the amount cannot be sent to the
// module account.
func (k Keeper) FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error {
	// if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amount); err != nil {
	// 	return err
	// }

	// feePool, err := k.FeePool.Get(ctx)
	// if err != nil {
	// 	return err
	// }

	// feePool.CommunityPool = feePool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(amount...)...)
	// return k.FeePool.Set(ctx, feePool)

	// since CommunityPool has a separate module account, send funds directly to its account
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amount)
}

// DistributeFromFeePool distributes funds from the pool module account to
// a receiver address while updating the community pool
func (k Keeper) DistributeFromFeePool(ctx context.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error {
	// feePool, err := k.FeePool.Get(ctx)
	// if err != nil {
	// 	return err
	// }

	// // NOTE the community pool isn't a module account, however its coins
	// // are held in the distribution module account. Thus the community pool
	// // must be reduced separately from the SendCoinsFromModuleToAccount call
	// newPool, negative := feePool.CommunityPool.SafeSub(sdk.NewDecCoinsFromCoins(amount...))
	// if negative {
	// 	return types.ErrBadDistribution
	// }

	// feePool.CommunityPool = newPool

	// err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiveAddr, amount)
	// if err != nil {
	// 	return err
	// }

	// return k.FeePool.Set(ctx, feePool)

	// since community pool is a module account and coins are held there, distribute funds from there
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiveAddr, amount)
}
