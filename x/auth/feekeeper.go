package auth

import (
	"fmt"
	
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	collectedFeesKey = []byte("collectedFees")
)

// FeeCollectionKeeper handles collection of fees in the anteHandler
// and setting of MinFees for different fee tokens
type FeeCollectionKeeper struct {

	// The (unexposed) key used to access the fee store from the Context.
	key sdk.StoreKey

	// The codec codec for binary encoding/decoding of accounts.
	cdc *codec.Codec
}

// NewFeeCollectionKeeper returns a new FeeCollectionKeeper
func NewFeeCollectionKeeper(cdc *codec.Codec, key sdk.StoreKey) FeeCollectionKeeper {
	return FeeCollectionKeeper{
		key: key,
		cdc: cdc,
	}
}

// GetCollectedFees - retrieves the collected fee pool
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

// AddCollectedFees - add to the fee pool
func (fck FeeCollectionKeeper) AddCollectedFees(ctx sdk.Context, coins sdk.Coins) sdk.Coins {
	logger := ctx.Logger().With("module", "auth")
	oldCoins := fck.GetCollectedFees(ctx)
	newCoins := oldCoins.Add(coins)
	fck.setCollectedFees(ctx, newCoins)
	logger.Debug(fmt.Sprintf("collect fee to pool, oldCoins: %v, addCoins: %v, newCoins: %v",
		oldCoins, coins, newCoins))
	return newCoins
}

// SubCollectedFees - sub fee from fee pool
func (fck FeeCollectionKeeper) SubCollectedFees(ctx sdk.Context, coins sdk.Coins) sdk.Coins {
	logger := ctx.Logger().With("module", "auth")
	oldCoins := fck.GetCollectedFees(ctx)
	newCoins, anyNeg := oldCoins.SafeSub(coins)
	if !anyNeg {
		fck.setCollectedFees(ctx, newCoins)
		logger.Debug(fmt.Sprintf("sub fee from pool, oldCoins: %v, subCoins: %v, newCoins: %v",
			oldCoins, coins, newCoins))
	} else {
		logger.Error(fmt.Sprintf("sub fee from pool failed, oldCoins: %v, subCoins: %v",
			oldCoins, coins))
	}
	
	return newCoins
}

// ClearCollectedFees - clear the fee pool
func (fck FeeCollectionKeeper) ClearCollectedFees(ctx sdk.Context) {
	fck.setCollectedFees(ctx, sdk.Coins{})
}
