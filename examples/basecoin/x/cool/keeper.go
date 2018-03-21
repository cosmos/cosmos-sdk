package cool

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// Keeper - handlers sets/gets of custom variables for your module
type Keeper struct {
	ck bank.CoinKeeper

	storeKey sdk.StoreKey // The (unexposed) key used to access the store from the Context.
}

// NewKeeper - Returns the Keeper
func NewKeeper(key sdk.StoreKey, bankKeeper bank.CoinKeeper) Keeper {
	return Keeper{bankKeeper, key}
}

// Key to knowing the trend on the streets!
var trendKey = []byte("TrendKey")

// GetTrend - returns the current cool trend
func (k Keeper) GetTrend(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(trendKey)
	return string(bz)
}

// Implements sdk.AccountMapper.
func (k Keeper) setTrend(ctx sdk.Context, newTrend string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(trendKey, []byte(newTrend))
}

// CheckTrend - Returns true or false based on whether guessedTrend is currently cool or not
func (k Keeper) CheckTrend(ctx sdk.Context, guessedTrend string) bool {
	if guessedTrend == k.GetTrend(ctx) {
		return true
	}
	return false
}
