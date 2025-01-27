package coretesting

import (
	"context"

	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
)

type dummyKey struct{}

var _ context.Context = &TestContext{}

type TestContext struct {
	context.Context
}

func Context() TestContext {
	dummy := &dummyCtx{
		stores:      map[string]store.KVStore{},
		events:      map[string][]event.Event{},
		protoEvents: map[string][]transaction.Msg{},
		header:      header.Info{},
		execMode:    transaction.ExecModeFinalize,
		gasConfig:   gas.GasConfig{},
		gasMeter:    nil,
	}

	return TestContext{
		Context: context.WithValue(context.Background(), dummyKey{}, dummy),
	}
}

// WithHeaderInfo sets the header on a testing ctx and returns the updated ctx.
func (t TestContext) WithHeaderInfo(info header.Info) TestContext {
	dummy := unwrap(t.Context)
	dummy.header = info

	return TestContext{
		Context: context.WithValue(t.Context, dummyKey{}, dummy),
	}
}

// WithExecMode sets the exec mode on a testing ctx and returns the updated ctx.
func (t TestContext) WithExecMode(mode transaction.ExecMode) TestContext {
	dummy := unwrap(t.Context)
	dummy.execMode = mode

	return TestContext{
		Context: context.WithValue(t.Context, dummyKey{}, dummy),
	}
}

// WithGas sets the gas config and meter on a testing ctx and returns the updated ctx.
func (t TestContext) WithGas(gasConfig gas.GasConfig, gasMeter gas.Meter) TestContext {
	dummy := unwrap(t.Context)
	dummy.gasConfig = gasConfig
	dummy.gasMeter = gasMeter

	return TestContext{
		Context: context.WithValue(t.Context, dummyKey{}, dummy),
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
