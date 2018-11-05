package auth

import (
	codec "github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	collectedFeesKey = []byte("collectedFees")
)

// This FeeCollectionKeeper handles collection of fees in the anteHandler
// and setting of MinFees for different fee tokens
type FeeCollectionKeeper struct {

	// The (unexposed) key used to access the fee store from the Context.
	key sdk.StoreKey

	// The codec codec for binary encoding/decoding of accounts.
	cdc *codec.Codec
}

func NewFeeCollectionKeeper(cdc *codec.Codec, key sdk.StoreKey) FeeCollectionKeeper {
	return FeeCollectionKeeper{
		key: key,
		cdc: cdc,
	}
}

// retrieves the collected fee pool
func (fck FeeCollectionKeeper) GetCollectedFees(ctx sdk.Context) sdk.Coins {
	store := ctx.KVStore(fck.key)
	bz := store.Get(collectedFeesKey)
	if bz == nil {
		return sdk.Coins{}
	}

	feePool := &(sdk.Coins{})
	fck.cdc.MustUnmarshalBinaryLengthPrefixed(bz, feePool)
	return *feePool
}

func (fck FeeCollectionKeeper) setCollectedFees(ctx sdk.Context, coins sdk.Coins) {
	bz := fck.cdc.MustMarshalBinaryLengthPrefixed(coins)
	store := ctx.KVStore(fck.key)
	store.Set(collectedFeesKey, bz)
}

// add to the fee pool
func (fck FeeCollectionKeeper) AddCollectedFees(ctx sdk.Context, coins sdk.Coins) sdk.Coins {
	newCoins := fck.GetCollectedFees(ctx).Plus(coins)
	fck.setCollectedFees(ctx, newCoins)

	return newCoins
}

// clear the fee pool
func (fck FeeCollectionKeeper) ClearCollectedFees(ctx sdk.Context) {
	fck.setCollectedFees(ctx, sdk.Coins{})
}
