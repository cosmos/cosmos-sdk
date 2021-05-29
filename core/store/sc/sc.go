package sc

import (
	"context"

	"github.com/cosmos/cosmos-sdk/core/container"
	"github.com/cosmos/cosmos-sdk/core/store"
)

type StoreKey struct{}

func StoreKeyProvider(scope container.Scope) StoreKey {
	panic("TODO")
}

func (StoreKey) Open(context.Context) store.BasicKVStore {
	panic("TODO")
}
