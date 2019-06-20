package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/uniswap/internal/types"

	"github.com/tendermint/tendermint/libs/log"
)

// Uniswap Keeper
type Keeper struct {
	// The key used to access the uniswap store
	storeKey sdk.StoreKey

	// The reference to the BankKeeper to retrieve account balances
	bk types.BankKeeper

	// The reference to the SupplyKeeper to hold coins for this module
	sk types.SupplyKeeper

	// The codec codec for binary encoding/decoding.
	cdc *codec.Codec

	// The reference to the Paramstore to get and set uniswap specific params
	paramSpace params.Subspace
}

// NewKeeper returns a uniswap keeper. It handles:
// - creating new reserve pools
// - facilitating swaps
// - users adding liquidity to a reserve pool
// - users removing liquidity from a reserve pool
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, bk types.BankKeeper, sk types.SupplyKeeper, paramSpace params.Subspace) Keeper {
	return Keeper{
		storeKey:   key,
		bk:         bk,
		sk:         sk,
		cdc:        cdc,
		paramSpace: paramSpace.WithKeyTable(types.ParamKeyTable()),
	}
}

// Logger returns a module-specific logger.
func (keeper Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}

// CreateReservePool initializes a new reserve pool for the new denomination
func (keeper Keeper) CreateReservePool(ctx sdk.Context, denom string) {
	store := ctx.KVStore(keeper.storeKey)
	key := GetReservePoolKey(denom)
	bz := store.Get(key)
	if bz != nil {
		panic(fmt.Sprintf("reserve pool for %s already exists", denom))
	}

	store.Set(key, keeper.cdc.MustMarshalBinaryBare(sdk.ZeroInt()))
}

// GetUNIForAddress returns the total UNI at the provided address
func (keeper Keeper) GetUNIForAddress(ctx sdk.Context, addr sdk.AccAddress) (balance sdk.Int) {
	store := ctx.KVStore(keeper.storeKey)
	key := GetUNIBalancesKey(addr)
	bz := store.Get(key)
	if bz != nil {
		keeper.cdc.MustUnmarshalBinaryBare(bz, &balance)
	}

	return
}

// SetUNIForAddress sets the provided UNI at the given address
func (keeper Keeper) SetUNIForAddress(ctx sdk.Context, amt sdk.Int, addr sdk.AccAddress) {
	store := ctx.KVStore(keeper.storeKey)
	key := GetUNIBalancesKey(addr)
	store.Set(key, keeper.cdc.MustMarshalBinaryBare(amt))
}

// GetTotalUNI returns the total UNI currently in existence
func (keeper Keeper) GetTotalUNI(ctx sdk.Context) (totalUNI sdk.Int) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(TotalUNIKey)
	if bz == nil {
		return sdk.ZeroInt()
	}

	keeper.cdc.MustUnmarshalBinaryBare(bz, &totalUNI)
	return
}

// SetTotalLiquidity sets the total UNI
func (keeper Keeper) SetTotalUNI(ctx sdk.Context, totalUNI sdk.Int) {
	store := ctx.KVStore(keeper.storeKey)
	store.Set(TotalUNIKey, keeper.cdc.MustMarshalBinaryBare(totalUNI))
}

// GetReservePool returns the total balance of an reserve pool at the provided denomination
func (keeper Keeper) GetReservePool(ctx sdk.Context, denom string) (balance sdk.Int) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(GetReservePoolKey(denom))
	if bz == nil {
		panic(fmt.Sprintf("reserve pool for %s does not exist", denom))
	}

	keeper.cdc.MustUnmarshalBinaryBare(bz, &balance)
	return
}

// GetNativeDenom returns the native denomination for this module from the global param store
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
