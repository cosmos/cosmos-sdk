package indexes

import (
	"context"

	"cosmossdk.io/core/coretesting"
	"cosmossdk.io/core/store"
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
