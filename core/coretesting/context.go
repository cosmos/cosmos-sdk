package coretesting

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/core/store"
)

func Context() context.Context {
	return &dummyCtx{
		stores: map[string]store.KVStore{},
	}
}

type dummyCtx struct {
	stores map[string]store.KVStore
}

func (d dummyCtx) Deadline() (deadline time.Time, ok bool) {
	panic("Deadline on dummy context")
}

func (d dummyCtx) Done() <-chan struct{} {
	panic("Done on dummy context")
}

func (d dummyCtx) Err() error {
	panic("Err on dummy context")
}

func (d dummyCtx) Value(key any) any {
	panic("Value on dummy context")

}

func unwrap(ctx context.Context) *dummyCtx {
	dummy, ok := ctx.(*dummyCtx)
	if !ok {
		panic(fmt.Sprintf("invalid context.Context, wanted dummy got: %T", ctx))
	}
	return dummy
}
