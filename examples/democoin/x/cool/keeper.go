package cool

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// Keeper - handlers sets/gets of custom variables for your module
type Keeper struct {
	ck bank.Keeper

	storeKey sdk.StoreKey // The (unexposed) key used to access the store from the Context.

	codespace sdk.CodespaceType
}

// NewKeeper - Returns the Keeper
func NewKeeper(key sdk.StoreKey, bankKeeper bank.Keeper, codespace sdk.CodespaceType) Keeper {
	return Keeper{bankKeeper, key, codespace}
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

// InitGenesis - store the genesis trend
func InitGenesis(ctx sdk.Context, k Keeper, data Genesis) error {
	k.setTrend(ctx, data.Trend)
	return nil
}

// WriteGenesis - output the genesis trend
func WriteGenesis(ctx sdk.Context, k Keeper) Genesis {
	trend := k.GetTrend(ctx)
	return Genesis{trend}
}
