package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/coinswap/internal/types"
	"github.com/cosmos/cosmos-sdk/x/params"

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

// TODO: rewrite
// CreateReservePool initializes a new reserve pool for the new denomination.
func (keeper Keeper) CreateReservePool(ctx sdk.Context, denom string) {
	store := ctx.KVStore(keeper.storeKey)
	key := GetReservePoolKey(denom)
	bz := store.Get(key)
	if bz != nil {
		panic(fmt.Sprintf("reserve pool for %s already exists", denom))
	}

	store.Set(key, keeper.cdc.MustMarshalBinaryBare(sdk.ZeroInt()))
}

// HasCoins returns whether or not an account has at least coins.
func (keeper Keeper) HasCoins(ctx sdk.Context, addr sdk.AccAddress, coins ...sdk.Coin) bool {
	return keeper.bk.HasCoins(ctx, addr, coins)
}

// BurnCoins burns liquidity coins from the ModuleAccount at moduleName.
func (keeper Keeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Int) {
	err := keeper.sk.BurnCoins(ctx, moduleName, sdk.NewCoin(moduleName, amt))
	if err != nil {
		panic(err)
	}
}

// MintCoins mints liquidity coins to the ModuleAccount at moduleName.
func (keeper Keeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) {
	err := keeper.sk.MintCoins(ctx, moduleName, amt)
	if err != nil {
		panic(err)
	}
}

// SendCoin sends coins from the address to the ModuleAccount at moduleName.
func (keeper Keeper) SendCoins(ctx sdk.Context, addr sdk.AccAddress, moduleName string, coins ...sdk.Coin) {
	err := keeper.sk.SendCoinsFromAccountToModule(ctx, msg.Sender, moduleName, coins)
	if err != nil {
		panic(err)
	}
}

// RecieveCoin sends coins from the ModuleAccount at moduleName to the
// address provided.
func (keeper Keeper) RecieveCoins(ctx sdk.Context, addr sdk.AccAddress, moduleName string, coin ...sdk.Coin) {
	err := keeper.sk.SendCoinsFromModuleToAccount(ctx, moduleName, msg.Sender, coins)
	if err != nil {
		panic(err)
	}
}

// TODO: rewrite
// GetReservePool returns the total balance of an reserve pool at the provided denomination.
func (keeper Keeper) GetReservePool(ctx sdk.Context, denom string) (coins sdk.Coins, found bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(GetReservePoolKey(denom))
	if bz == nil {
		return
	}

	keeper.cdc.MustUnmarshalBinaryBare(bz, &balance)
	return balance, true
}

// GetNativeDenom returns the native denomination for this module from the
// global param store.
func (keeper Keeper) GetNativeDenom(ctx sdk.Context) (nativeDenom string) {
	keeper.paramSpace.Get(ctx, types.KeyNativeDenom, &nativeDenom)
	return
}

// GetFeeParam returns the current FeeParam from the global param store
func (keeper Keeper) GetFeeParam(ctx sdk.Context) (feeParams sdk.Dec) {
	keeper.paramSpace.Get(ctx, types.KeyFee, &feeParams)
	return
}

func (keeper Keeper) SetFeeParam(ctx sdk.Context, feeParams sdk.Dec) {
	keeper.paramSpace.Set(ctx, types.KeyFee, &feeParams)
}

// Logger returns a module-specific logger.
func (keeper Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
