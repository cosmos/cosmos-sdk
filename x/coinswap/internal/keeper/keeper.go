package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/coinswap/internal/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	supply "github.com/cosmos/cosmos-sdk/x/supply/types"

	"github.com/tendermint/tendermint/libs/log"
)

// Keeper of the coinswap store
type Keeper struct {
	cdc        *codec.Codec
	storeKey   sdk.StoreKey
	bk         types.BankKeeper
	sk         types.SupplyKeeper
	paramSpace params.Subspace
}

// NewKeeper returns a coinswap keeper. It handles:
// - creating new ModuleAccounts for each trading pair
// - burning minting liquidity coins
// - sending to and from ModuleAccounts
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, bk types.BankKeeper, sk types.SupplyKeeper, paramSpace params.Subspace) Keeper {
	return Keeper{
		storeKey:   key,
		bk:         bk,
		sk:         sk,
		cdc:        cdc,
		paramSpace: paramSpace.WithKeyTable(types.ParamKeyTable()),
	}
}

// CreateReservePool initializes a new reserve pool by creating a
// ModuleAccount with minting and burning permissions.
func (keeper Keeper) CreateReservePool(ctx sdk.Context, moduleName string) {
	moduleAcc := keeper.sk.GetModuleAccount(ctx, moduleName)
	if moduleAcc != nil {
		panic(fmt.Sprintf("reserve pool for %s already exists", moduleName))
	}
	// TODO: add burning permissions
	moduleAcc = supply.NewEmptyModuleAccount(moduleName, "minter")
	keeper.sk.SetModuleAccount(ctx, moduleAcc)
}

// HasCoins returns whether or not an account has at least coins.
func (keeper Keeper) HasCoins(ctx sdk.Context, addr sdk.AccAddress, coins ...sdk.Coin) bool {
	return keeper.bk.HasCoins(ctx, addr, coins)
}

// BurnCoins burns liquidity coins from the ModuleAccount at moduleName. The
// moduleName and denomination of the liquidity coins are the same.
func (keeper Keeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Int) {
	err := keeper.sk.BurnCoins(ctx, moduleName, sdk.NewCoins(sdk.NewCoin(moduleName, amt)))
	if err != nil {
		panic(err)
	}
}

// MintCoins mints liquidity coins to the ModuleAccount at moduleName. The
// moduleName and denomination of the liquidity coins are the same.
func (keeper Keeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Int) {
	err := keeper.sk.MintCoins(ctx, moduleName, sdk.NewCoins(sdk.NewCoin(moduleName, amt)))
	if err != nil {
		panic(err)
	}
}

// SendCoin sends coins from the address to the ModuleAccount at moduleName.
func (keeper Keeper) SendCoins(ctx sdk.Context, addr sdk.AccAddress, moduleName string, coins ...sdk.Coin) {
	err := keeper.sk.SendCoinsFromAccountToModule(ctx, addr, moduleName, coins)
	if err != nil {
		panic(err)
	}
}

// RecieveCoin sends coins from the ModuleAccount at moduleName to the
// address provided.
func (keeper Keeper) RecieveCoins(ctx sdk.Context, addr sdk.AccAddress, moduleName string, coins ...sdk.Coin) {
	err := keeper.sk.SendCoinsFromModuleToAccount(ctx, moduleName, addr, coins)
	if err != nil {
		panic(err)
	}
}

// GetReservePool returns the total balance of an reserve pool at the
// provided denomination.
func (keeper Keeper) GetReservePool(ctx sdk.Context, moduleName string) (coins sdk.Coins, found bool) {
	acc := keeper.sk.GetModuleAccount(ctx, moduleName)
	if acc != nil {
		return acc.GetCoins(), true
	}
	return coins, false
}

// GetNativeDenom returns the native denomination for this module from the
// global param store.
func (keeper Keeper) GetNativeDenom(ctx sdk.Context) (nativeDenom string) {
	keeper.paramSpace.Get(ctx, types.KeyNativeDenom, &nativeDenom)
	return
}

// GetFeeParam returns the current FeeParam from the global param store
func (keeper Keeper) GetFeeParam(ctx sdk.Context) (feeParam types.FeeParam) {
	keeper.paramSpace.Get(ctx, types.KeyFee, &feeParam)
	return
}

// SetParams sets the parameters for the coinswap module.
func (keeper Keeper) SetParams(ctx sdk.Context, params types.Params) {
	keeper.paramSpace.SetParamSet(ctx, &params)
}

// Logger returns a module-specific logger.
func (keeper Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
