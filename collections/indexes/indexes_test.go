package indexes

import (
	"context"

	store "cosmossdk.io/collections/corecompat"
	"cosmossdk.io/collections/internal/testutil"
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
