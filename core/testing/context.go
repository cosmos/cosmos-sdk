package coretesting

import (
	"context"
	"time"

	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
)

type dummyKey struct{}

var _ context.Context = &TestContext{}

type TestContext struct {
	ctx context.Context
}

func Context() TestContext {
	dummy := &dummyCtx{
		stores:      map[string]store.KVStore{},
		events:      map[string][]event.Event{},
		protoEvents: map[string][]transaction.Msg{},
		header:      header.Info{},
		execMode:    transaction.ExecModeFinalize,
	}

	return TestContext{
		ctx: context.WithValue(context.Background(), dummyKey{}, dummy),
	}
}

func (t TestContext) Deadline() (deadline time.Time, ok bool) {
	return t.ctx.Deadline()
}

func (t TestContext) Done() <-chan struct{} {
	return t.ctx.Done()
}

func (t TestContext) Err() error {
	return t.ctx.Err()
}

func (t TestContext) Value(key any) any {
	return t.ctx.Value(key)
}

// WithHeaderInfo sets the header on a testing ctx and returns the updated ctx.
func (t TestContext) WithHeaderInfo(info header.Info) TestContext {
	dummy := unwrap(t.ctx)
	dummy.header = info

	return TestContext{
		ctx: context.WithValue(t.ctx, dummyKey{}, dummy),
	}
}

// WithExecMode sets the exec mode on a testing ctx and returns the updated ctx.
func (t TestContext) WithExecMode(mode transaction.ExecMode) context.Context {
	dummy := unwrap(t.ctx)
	dummy.execMode = mode

	return TestContext{
		ctx: context.WithValue(t.ctx, dummyKey{}, dummy),
	}
}

// WithGas sets the gas config and meter on a testing ctx and returns the updated ctx.
func (t TestContext) WithGas(gasConfig gas.GasConfig, gasMeter gas.Meter) context.Context {
	dummy := unwrap(t.ctx)
	dummy.gasConfig = gasConfig
	dummy.gasMeter = gasMeter

	return TestContext{
		ctx: context.WithValue(t.ctx, dummyKey{}, dummy),
	}
}

type dummyCtx struct {
	// maps store by the actor.
	stores map[string]store.KVStore
	// maps event emitted by the actor.
	events map[string][]event.Event
	// maps proto events emitted by the actor.
	protoEvents map[string][]transaction.Msg

	header   header.Info
	execMode transaction.ExecMode

	gasMeter  gas.Meter
	gasConfig gas.GasConfig
}

func unwrap(ctx context.Context) *dummyCtx {
	dummy := ctx.Value(dummyKey{})
	if dummy == nil {
		panic("invalid ctx without dummy")
	}

	return dummy.(*dummyCtx)
}
