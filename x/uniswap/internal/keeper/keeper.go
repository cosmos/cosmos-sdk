package keeper

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	supply "github.com/cosmos/cosmos-sdk/x/supply/keeper"
	"github.com/cosmos/cosmos-sdk/x/uniswap/internal/types"

	"github.com/tendermint/tendermint/libs/log"
)

// Uniswap Keeper
type Keeper struct {
	// The key used to access the uniswap store
	storeKey sdk.StoreKey

	// The reference to the SupplyKeeper to hold coins for this module
	types.SupplyKeeper

	// The codec codec for binary encoding/decoding.
	cdc *codec.Codec

	// The reference to the Paramstore to get and set uniswap specific params
	paramSpace params.Subspace
}

// NewKeeper returns a uniswap keeper. It handles:
// - creating new exchanges
// - facilitating swaps
// - users adding liquidity to exchanges
// - users removing liquidity to exchanges
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, supplyKeeper supply.Keeper, paramSpace params.Subspace) Keeper {
	return Keeper{
		storeKey:     key,
		supplyKeeper: supplyKeeper,
		cdc:          cdc,
		paramSpace:   paramSpace.WithKeyTable(types.ParamKeyTable()),
	}
}

// Logger returns a module-specific logger.
func (keeper Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}

// CreateExchange initializes a new exchange pair between the new coin and the native asset
func (keeper Keeper) CreateExchange(ctx sdk.Context, exchangeDenom string) {
	store := ctx.KVStore(keeper.storeKey)
	key := GetExchangeKey(exchangeDenom)
	bz := store.Get(key)
	if bz != nil {
		panic("exchange already exists")
	}

	store.Set(key, keeper.encode(sdk.ZeroInt()))
}

// GetUNIForAddress returns the total UNI at the provided address
func (keeper Keeper) GetUNIForAddress(ctx sdk.Context, addr sdk.AccAddress) sdk.Int {
	var balance sdk.Int

	store := ctx.KVStore(keeper.storeKey)
	key := GetUNIBalancesKey(addr)
	bz := store.Get(key)
	if bz != nil {
		balance = keeper.decode(bz)
	}

	return balance
}

// SetUNIForAddress sets the provided UNI at the given address
func (keeper Keeper) SetUNIForAddress(ctx sdk.Context, amt sdk.Int, addr sdk.AccAddress) {
	store := ctx.KVStore(keeper.storeKey)
	key := GetUNIBalancesKey(addr)
	store.Set(key, keeper.encode(amt))
}

// GetTotalUNI returns the total UNI currently in existence
func (keeper Keeper) GetTotalUNI(ctx sdk.Context) sdk.Int {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(TotalUNIKey)
	if bz == nil {
		return sdk.ZeroInt()
	}

	totalUNI := keeper.decode(bz)
	return totalUNI
}

// SetTotalLiquidity sets the total UNI
func (keeper Keeper) SetTotalUNI(ctx sdk.Context, totalUNI sdk.Int) {
	store := ctx.KVStore(keeper.storeKey)
	store.Set(TotalUNIKey, keeper.encode(totalUNI))
}

// GetExchange returns the total balance of an exchange at the provided denomination
func (keeper Keeper) GetExchange(ctx sdk.Context, denom string) sdk.Int {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(GetExchangeKey(denom))
	if bz == nil {
		panic(fmt.Sprintf("exchange for denomination: %s does not exist"), denom)
	}
	return keeper.decode(bz)
}

// GetFeeParams returns the current FeeParams from the global param store
func (keeper Keeper) GetFeeParams(ctx sdk.Context) (feeParams types.FeeParams) {
	keeper.paramSpace.Get(ctx, types.ParamStoreKeyFeeParams, &feeParams)
	return feeParams
}

func (keeper Keeper) setFeeParams(ctx sdk.Context, feeParams types.FeeParams) {
	keeper.paramSpace.Set(ctx, types.ParamStoreKeyFeeParams, &feeParams)
}

// -----------------------------------------------------------------------------
// Misc.

func (keeper Keeper) decode(bz []byte) (balance sdk.Int) {
	err := keeper.cdc.UnmarshalBinaryBare(bz, &balance)
	if err != nil {
		panic(err)
	}
	return
}

func (keeper Keeper) encode(balance sdk.Int) (bz []byte) {
	bz, err := keeper.cdc.MarshalBinaryBare(balance)
	if err != nil {
		panic(err)
	}
	return
}
