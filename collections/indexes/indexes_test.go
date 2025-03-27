package indexes

import (
	"context"

	"cosmossdk.io/collections/internal/testutil"

	store "cosmossdk.io/collections/corecompat"
)

func deps() (store.KVStoreService, context.Context) {
	ctx := testutil.Context()
	kv := testutil.KVStoreService(ctx, "test")
	return kv, ctx
}

type company struct {
	City string
	Vat  uint64
}
