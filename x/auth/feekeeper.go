package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

// This FeeCollectionKeeper handles collection of fees in the anteHandler
// and setting of MinFees for different fee tokens
type FeeCollectionKeeper struct {

	// The (unexposed) key used to access the fee store from the Context.
	key sdk.StoreKey

	// The wire codec for binary encoding/decoding of accounts.
	cdc *wire.Codec
}

// NewFeeKeeper returns a new FeeKeeper
func NewFeeCollectionKeeper(cdc *wire.Codec, key sdk.StoreKey) FeeCollectionKeeper {
	return FeeCollectionKeeper{
		key: key,
		cdc: cdc,
	}
}

// Adds to Collected Fee Pool
func (fck FeeCollectionKeeper) GetCollectedFees(ctx sdk.Context) sdk.Coins {
	store := ctx.KVStore(fck.key)
	bz := store.Get([]byte("collectedFees"))
	if bz == nil {
		return sdk.Coins{}
	}

	feePool := &(sdk.Coins{})
	err := fck.cdc.UnmarshalBinary(bz, feePool)
	if err != nil {
		panic("should not happen")
	}
	return *feePool
}

// Sets to Collected Fee Pool
func (fck FeeCollectionKeeper) setCollectedFees(ctx sdk.Context, coins sdk.Coins) {
	bz, err := fck.cdc.MarshalBinary(coins)
	if err != nil {
		panic("should not happen")
	}

	store := ctx.KVStore(fck.key)
	store.Set([]byte("collectedFees"), bz)
}

// Adds to Collected Fee Pool
func (fck FeeCollectionKeeper) addCollectedFees(ctx sdk.Context, coins sdk.Coins) sdk.Coins {
	newCoins := fck.GetCollectedFees(ctx).Plus(coins)
	fck.setCollectedFees(ctx, newCoins)

	return newCoins
}

// Clears the collected Fee Pool
func (fck FeeCollectionKeeper) ClearCollectedFees(ctx sdk.Context) {
	fck.setCollectedFees(ctx, sdk.Coins{})
}
