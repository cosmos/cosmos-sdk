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

// Key to knowing whats cool
var whatCoolKey = []byte("WhatsCoolKey")

// Implements sdk.AccountMapper.
func (am Mapper) GetWhatCool(ctx sdk.Context) string {
	store := ctx.KVStore(am.key)
	bz := store.Get(whatCoolKey)
	return string(bz)
}

// Implements sdk.AccountMapper.
func (am Mapper) SetWhatCool(ctx sdk.Context, whatscool string) {
	store := ctx.KVStore(am.key)
	store.Set(whatCoolKey, []byte(whatscool))
}
