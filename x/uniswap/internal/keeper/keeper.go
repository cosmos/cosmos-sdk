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
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, bk types.BankKeeper, sk supply.Keeper, paramSpace params.Subspace) Keeper {
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

// GetReservePool returns the total balance of an reserve pool at the provided denomination
func (keeper Keeper) GetReservePool(ctx sdk.Context, denom string) sdk.Int {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(GetReservePoolKey(denom))
	if bz == nil {
		panic(fmt.Sprintf("reserve pool for %s does not exist"), denom)
	}
	return keeper.decode(bz)
}

// GetFeeParams returns the current FeeParams from the global param store
func (keeper Keeper) GetFeeParams(ctx sdk.Context) (feeParams types.FeeParams) {
	keeper.paramSpace.Get(ctx, types.ParamStoreKeyFeeParams, &feeParams)
	return feeParams
}

func (keeper Keeper) SetFeeParams(ctx sdk.Context, feeParams types.FeeParams) {
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
