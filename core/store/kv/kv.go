package kv

import (
	"context"

	"github.com/cosmos/cosmos-sdk/core/container"
	"github.com/cosmos/cosmos-sdk/core/store"
	"github.com/cosmos/cosmos-sdk/core/store/sc"
	"github.com/cosmos/cosmos-sdk/core/store/ss"
)

type StoreKey struct{}

func StoreKeyProvider(scope container.Scope, ssKey ss.StoreKey, scKey sc.StoreKey) StoreKey {
	panic("TODO")
}

func (StoreKey) Open(context.Context) store.KVStore {
	panic("TODO")
}
