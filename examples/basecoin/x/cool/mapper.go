package cool

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// This Cool Mapper handlers sets/gets of custom variables for your module
type Mapper struct {
	key sdk.StoreKey // The (unexposed) key used to access the store from the Context.
}

func NewMapper(key sdk.StoreKey) Mapper {
	return Mapper{key}
}

// Key to knowing the trend on the streets!
var trendKey = []byte("TrendKey")

// Implements sdk.AccountMapper.
func (am Mapper) GetTrend(ctx sdk.Context) string {
	store := ctx.KVStore(am.key)
	bz := store.Get(trendKey)
	return string(bz)
}

// Implements sdk.AccountMapper.
func (am Mapper) SetTrend(ctx sdk.Context, newTrend string) {
	store := ctx.KVStore(am.key)
	store.Set(trendKey, []byte(newTrend))
}
