package indexes

import (
	"context"

	"cosmossdk.io/core/testing"

	store "cosmossdk.io/collections/corecompat"
)

func deps() (store.KVStoreService, context.Context) {
	ctx := coretesting.Context()
	kv := coretesting.KVStoreService(ctx, "test")
	return kv, ctx
}

type company struct {
	City string
	Vat  uint64
}
