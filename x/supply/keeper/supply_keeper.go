package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

// SupplyKeeper defines the keeper of the supply store
type SupplyKeeper struct {
	cdc      *codec.Codec
	storeKey sdk.StoreKey
}

// NewSupplyKeeper creates a new supply Keeper instance
func NewSupplyKeeper(cdc *codec.Codec, key sdk.StoreKey) SupplyKeeper {
	return SupplyKeeper{
		cdc:      cdc,
		storeKey: key,
	}
}

// GetSupply retrieves the Supply from store
func (sk SupplyKeeper) GetSupply(ctx sdk.Context) (supply types.Supply) {
	store := ctx.KVStore(sk.storeKey)
	b := store.Get(supplyKey)
	if b == nil {
		panic("Stored supply should not have been nil")
	}
	sk.cdc.MustUnmarshalBinaryLengthPrefixed(b, &supply)
	return
}

// SetSupply sets the Supply to store
func (sk SupplyKeeper) SetSupply(ctx sdk.Context, supply types.Supply) {
	store := ctx.KVStore(sk.storeKey)
	b := sk.cdc.MustMarshalBinaryLengthPrefixed(supply)
	store.Set(supplyKey, b)
}

// Inflate increases the total supply amount
func (sk SupplyKeeper) Inflate(ctx sdk.Context, amount sdk.Coins) {
	supply := sk.GetSupply(ctx)
	supply.Inflate(amount)
	sk.SetSupply(ctx, supply)
}

// Deflate reduces the total supply amount
func (sk SupplyKeeper) Deflate(ctx sdk.Context, amount sdk.Coins) {
	supply := sk.GetSupply(ctx)
	supply.Deflate(amount)
	sk.SetSupply(ctx, supply)
}
