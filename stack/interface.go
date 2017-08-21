//nolint
package stack

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

// Middleware is anything that wraps another handler to enhance functionality.
//
// You can use utilities in handlers to construct them, the interfaces
// are exposed in the top-level package to avoid import loops.
type Middleware interface {
	CheckerMiddle
	DeliverMiddle
	InitStaterMiddle
	InitValidaterMiddle
	sdk.Named
}

type CheckerMiddle interface {
	CheckTx(ctx sdk.Context, store state.SimpleDB,
		tx sdk.Tx, next sdk.Checker) (sdk.CheckResult, error)
}

type CheckerMiddleFunc func(sdk.Context, state.SimpleDB,
	sdk.Tx, sdk.Checker) (sdk.CheckResult, error)

func (c CheckerMiddleFunc) CheckTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, next sdk.Checker) (sdk.CheckResult, error) {
	return c(ctx, store, tx, next)
}

type DeliverMiddle interface {
	DeliverTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx,
		next sdk.Deliver) (sdk.DeliverResult, error)
}

type DeliverMiddleFunc func(sdk.Context, state.SimpleDB,
	sdk.Tx, sdk.Deliver) (sdk.DeliverResult, error)

func (d DeliverMiddleFunc) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, next sdk.Deliver) (sdk.DeliverResult, error) {
	return d(ctx, store, tx, next)
}

type InitStaterMiddle interface {
	InitState(l log.Logger, store state.SimpleDB, module,
		key, value string, next sdk.InitStater) (string, error)
}

type InitStaterMiddleFunc func(log.Logger, state.SimpleDB,
	string, string, string, sdk.InitStater) (string, error)

func (c InitStaterMiddleFunc) InitState(l log.Logger, store state.SimpleDB,
	module, key, value string, next sdk.InitStater) (string, error) {
	return c(l, store, module, key, value, next)
}

type InitValidaterMiddle interface {
	InitValidate(l log.Logger, store state.SimpleDB, vals []*abci.Validator, next sdk.InitValidater)
}

type InitValidaterMiddleFunc func(log.Logger, state.SimpleDB,
	[]*abci.Validator, sdk.InitValidater)

func (c InitValidaterMiddleFunc) InitValidate(l log.Logger, store state.SimpleDB,
	vals []*abci.Validator, next sdk.InitValidater) {
	c(l, store, vals, next)
}

// holders
type PassCheck struct{}

func (_ PassCheck) CheckTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, next sdk.Checker) (sdk.CheckResult, error) {
	return next.CheckTx(ctx, store, tx)
}

type PassDeliver struct{}

func (_ PassDeliver) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, next sdk.Deliver) (sdk.DeliverResult, error) {
	return next.DeliverTx(ctx, store, tx)
}

type PassInitState struct{}

func (_ PassInitState) InitState(l log.Logger, store state.SimpleDB, module,
	key, value string, next sdk.InitStater) (string, error) {
	return next.InitState(l, store, module, key, value)
}

type PassInitValidate struct{}

func (_ PassInitValidate) InitValidate(l log.Logger, store state.SimpleDB,
	vals []*abci.Validator, next sdk.InitValidater) {
	next.InitValidate(l, store, vals)
}

// Dispatchable is like middleware, except the meaning of "next" is different.
// Whereas in the middleware, it is the next handler that we should pass the same tx into,
// for dispatchers, it is a dispatcher, which it can use to
type Dispatchable interface {
	Middleware
	AssertDispatcher()
}

// WrapHandler turns a sdk.Handler into a Dispatchable interface
func WrapHandler(h sdk.Handler) Dispatchable {
	return wrapped{h}
}

type wrapped struct {
	h sdk.Handler
}

var _ Dispatchable = wrapped{}

func (w wrapped) AssertDispatcher() {}

func (w wrapped) Name() string {
	return w.h.Name()
}

func (w wrapped) CheckTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, _ sdk.Checker) (sdk.CheckResult, error) {
	return w.h.CheckTx(ctx, store, tx)
}

func (w wrapped) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, _ sdk.Deliver) (sdk.DeliverResult, error) {
	return w.h.DeliverTx(ctx, store, tx)
}

func (w wrapped) InitState(l log.Logger, store state.SimpleDB,
	module, key, value string, _ sdk.InitStater) (string, error) {
	return w.h.InitState(l, store, module, key, value)
}

func (w wrapped) InitValidate(l log.Logger, store state.SimpleDB,
	vals []*abci.Validator, next sdk.InitValidater) {
	w.h.InitValidate(l, store, vals)
}
