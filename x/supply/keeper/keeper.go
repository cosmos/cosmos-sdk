package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

// Keeper of the supply store
type Keeper struct {
	*bank.BaseViewKeeper

	cdc      *codec.Codec
	storeKey sdk.StoreKey
	ak       auth.AccountKeeper
}

// NewKeeper creates a new Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, ak auth.AccountKeeper, codespace sdk.CodespaceType) Keeper {
	baseViewKeeper := bank.NewBaseViewKeeper(ak, codespace)
	return Keeper{
		&baseViewKeeper,
		cdc,
		key,
		ak,
	}
}

// GetSupply retrieves the Supply from store
func (k Keeper) GetSupply(ctx sdk.Context) (supply types.Supply) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(supplyKey)
	if b == nil {
		panic("Stored supply should not have been nil")
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &supply)
	return
}

// SetSupply sets the Supply to store
func (k Keeper) SetSupply(ctx sdk.Context, supply types.Supply) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(supply)
	store.Set(supplyKey, b)
}
