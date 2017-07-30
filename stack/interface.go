//nolint
package stack

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/state"
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
	basecoin.Named
}

type CheckerMiddle interface {
	CheckTx(ctx basecoin.Context, store state.SimpleDB,
		tx basecoin.Tx, next basecoin.Checker) (basecoin.CheckResult, error)
}

type CheckerMiddleFunc func(basecoin.Context, state.SimpleDB,
	basecoin.Tx, basecoin.Checker) (basecoin.CheckResult, error)

func (c CheckerMiddleFunc) CheckTx(ctx basecoin.Context, store state.SimpleDB,
	tx basecoin.Tx, next basecoin.Checker) (basecoin.CheckResult, error) {
	return c(ctx, store, tx, next)
}

type DeliverMiddle interface {
	DeliverTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx,
		next basecoin.Deliver) (basecoin.DeliverResult, error)
}

type DeliverMiddleFunc func(basecoin.Context, state.SimpleDB,
	basecoin.Tx, basecoin.Deliver) (basecoin.DeliverResult, error)

func (d DeliverMiddleFunc) DeliverTx(ctx basecoin.Context, store state.SimpleDB,
	tx basecoin.Tx, next basecoin.Deliver) (basecoin.DeliverResult, error) {
	return d(ctx, store, tx, next)
}

type InitStaterMiddle interface {
	InitState(l log.Logger, store state.SimpleDB, module,
		key, value string, next basecoin.InitStater) (string, error)
}

type InitStaterMiddleFunc func(log.Logger, state.SimpleDB,
	string, string, string, basecoin.InitStater) (string, error)

func (c InitStaterMiddleFunc) InitState(l log.Logger, store state.SimpleDB,
	module, key, value string, next basecoin.InitStater) (string, error) {
	return c(l, store, module, key, value, next)
}

type InitValidaterMiddle interface {
	InitValidate(l log.Logger, store state.SimpleDB, vals []*abci.Validator, next basecoin.InitValidater)
}

type InitValidaterMiddleFunc func(log.Logger, state.SimpleDB,
	[]*abci.Validator, basecoin.InitValidater)

func (c InitValidaterMiddleFunc) InitValidate(l log.Logger, store state.SimpleDB,
	vals []*abci.Validator, next basecoin.InitValidater) {
	c(l, store, vals, next)
}

// holders
type PassCheck struct{}

func (_ PassCheck) CheckTx(ctx basecoin.Context, store state.SimpleDB,
	tx basecoin.Tx, next basecoin.Checker) (basecoin.CheckResult, error) {
	return next.CheckTx(ctx, store, tx)
}

type PassDeliver struct{}

func (_ PassDeliver) DeliverTx(ctx basecoin.Context, store state.SimpleDB,
	tx basecoin.Tx, next basecoin.Deliver) (basecoin.DeliverResult, error) {
	return next.DeliverTx(ctx, store, tx)
}

type PassInitState struct{}

func (_ PassInitState) InitState(l log.Logger, store state.SimpleDB, module,
	key, value string, next basecoin.InitStater) (string, error) {
	return next.InitState(l, store, module, key, value)
}

type PassInitValidate struct{}

func (_ PassInitValidate) InitValidate(l log.Logger, store state.SimpleDB,
	vals []*abci.Validator, next basecoin.InitValidater) {
	next.InitValidate(l, store, vals)
}

// Dispatchable is like middleware, except the meaning of "next" is different.
// Whereas in the middleware, it is the next handler that we should pass the same tx into,
// for dispatchers, it is a dispatcher, which it can use to
type Dispatchable interface {
	Middleware
	AssertDispatcher()
}

// WrapHandler turns a basecoin.Handler into a Dispatchable interface
func WrapHandler(h basecoin.Handler) Dispatchable {
	return wrapped{h}
}

type wrapped struct {
	h basecoin.Handler
}

var _ Dispatchable = wrapped{}

func (w wrapped) AssertDispatcher() {}

func (w wrapped) Name() string {
	return w.h.Name()
}

func (w wrapped) CheckTx(ctx basecoin.Context, store state.SimpleDB,
	tx basecoin.Tx, _ basecoin.Checker) (basecoin.CheckResult, error) {
	return w.h.CheckTx(ctx, store, tx)
}

func (w wrapped) DeliverTx(ctx basecoin.Context, store state.SimpleDB,
	tx basecoin.Tx, _ basecoin.Deliver) (basecoin.DeliverResult, error) {
	return w.h.DeliverTx(ctx, store, tx)
}

func (w wrapped) InitState(l log.Logger, store state.SimpleDB,
	module, key, value string, _ basecoin.InitStater) (string, error) {
	return w.h.InitState(l, store, module, key, value)
}

func (w wrapped) InitValidate(l log.Logger, store state.SimpleDB,
	vals []*abci.Validator, next basecoin.InitValidater) {
	w.h.InitValidate(l, store, vals)
}
