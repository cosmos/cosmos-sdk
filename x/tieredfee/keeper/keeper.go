package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/cosmos/cosmos-sdk/x/tieredfee/types"
)

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	paramSpace paramtypes.Subspace
}

// NewKeeper creates a Keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, paramSpace paramtypes.Subspace) Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc,
		storeKey,
		paramSpace,
	}
}

// SetBlockGasUsed record the block gas used at EndBlocker.
func (k Keeper) SetBlockGasUsed(ctx sdk.Context, gasUsed uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := sdk.Uint64ToBigEndian(gasUsed)
	store.Set(types.BlockGasUsedKey, bz)
}

// GetBlockGasUsed returns current recorded block gas used.
func (k Keeper) GetBlockGasUsed(ctx sdk.Context) (uint64, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.BlockGasUsedKey)
	if len(bz) == 0 {
		return 0, false
	}
	return sdk.BigEndianToUint64(bz), true
}

// SetGasPrice set the current gas price of the tier.
func (k Keeper) SetGasPrice(ctx sdk.Context, tier uint32, gasPrice sdk.DecCoins) {
	store := ctx.KVStore(k.storeKey)
	protoCoins := sdk.ProtoDecCoins{
		Coins: gasPrice,
	}
	bz := k.cdc.MustMarshal(&protoCoins)
	store.Set(types.GasPriceKey(tier), bz)
}

// GetGasPrice get the current gas price of the tier.
func (k Keeper) GetGasPrice(ctx sdk.Context, tier uint32) (sdk.DecCoins, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GasPriceKey(tier))
	if len(bz) == 0 {
		return nil, false
	}

	var protoCoins sdk.ProtoDecCoins
	err := k.cdc.Unmarshal(bz, &protoCoins)
	if err != nil {
		panic(err)
	}

	return protoCoins.Coins, true
}

// GetAllGasPrice get the gas prices for all tiers.
func (k Keeper) GetAllGasPrice(ctx sdk.Context) []sdk.DecCoins {
	params := k.GetParams(ctx)
	prices := make([]sdk.DecCoins, len(params.Tiers))
	for i, params := range params.Tiers {
		price, found := k.GetGasPrice(ctx, uint32(i))
		if !found {
			price = params.InitialGasPrice
		}
		prices[i] = price
	}
	return prices
}

// UpdateAllTiers update the gas prices for all tiers.
func (k Keeper) UpdateAllTiers(ctx sdk.Context) {
	params := k.GetParams(ctx)
	gasUsed, found := k.GetBlockGasUsed(ctx)
	if !found {
		// this could happen in the first block after the module is just enabled,
		// set to initialize_gas_price in this case.
		for i, tierParams := range params.Tiers {
			k.SetGasPrice(ctx, uint32(i), tierParams.InitialGasPrice)
		}
		return
	}

	for i, tierParams := range params.Tiers {
		gasPrice, found := k.GetGasPrice(ctx, uint32(i))
		if !found {
			gasPrice = tierParams.InitialGasPrice
		}
		newGasPrice := types.AdjustGasPrice(gasPrice, gasUsed, tierParams)
		k.SetGasPrice(ctx, uint32(i), newGasPrice)
	}
}
