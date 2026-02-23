package testutil

import (
	"context"

	"cosmossdk.io/collections/corecompat"
)

type dummyKey struct{}

func Context() context.Context {
	dummy := &dummyCtx{
		stores: map[string]corecompat.KVStore{},
	}

	ctx := context.WithValue(context.Background(), dummyKey{}, dummy)
	return ctx
}

type dummyCtx struct {
	// maps store by the actor.
	stores map[string]corecompat.KVStore
}

func unwrap(ctx context.Context) *dummyCtx {
	dummy := ctx.Value(dummyKey{})
	if dummy == nil {
		panic("invalid ctx without dummy")
	}

	return dummy.(*dummyCtx)
}
