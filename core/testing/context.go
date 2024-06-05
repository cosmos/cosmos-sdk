package coretesting

import (
	"context"

	"cosmossdk.io/core/store"
)

type dummyKey struct{}

func Context() context.Context {
	dummy := &dummyCtx{
		stores: map[string]store.KVStore{},
	}

	ctx := context.WithValue(context.Background(), dummyKey{}, dummy)
	return ctx
}

type dummyCtx struct {
	stores map[string]store.KVStore
}

func unwrap(ctx context.Context) *dummyCtx {
	dummy := ctx.Value(dummyKey{})
	if dummy == nil {
		panic("invalid ctx without dummy")
	}

	return dummy.(*dummyCtx)
}
