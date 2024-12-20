package coretesting

import (
	"context"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
)

type dummyKey struct{}

func Context() context.Context {
	dummy := &dummyCtx{
		stores:      map[string]store.KVStore{},
		events:      map[string][]event.Event{},
		protoEvents: map[string][]transaction.Msg{},
		header:      header.Info{},
	}

	ctx := context.WithValue(context.Background(), dummyKey{}, dummy)
	return ctx
}

// WithHeader sets the header on a testing ctx and returns the updated ctx.
func WithHeader(ctx context.Context, info header.Info) context.Context {
	dummy := unwrap(ctx)
	dummy.header = info

	return context.WithValue(ctx, dummyKey{}, dummy)
}

type dummyCtx struct {
	// maps store by the actor.
	stores map[string]store.KVStore
	// maps event emitted by the actor.
	events map[string][]event.Event
	// maps proto events emitted by the actor.
	protoEvents map[string][]transaction.Msg
	// header is the header set by a test runner
	header header.Info
}

func unwrap(ctx context.Context) *dummyCtx {
	dummy := ctx.Value(dummyKey{})
	if dummy == nil {
		panic("invalid ctx without dummy")
	}

	return dummy.(*dummyCtx)
}
