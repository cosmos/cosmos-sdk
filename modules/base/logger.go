package base

import (
	"time"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
)

// nolint
const (
	NameLogger = "lggr"
)

// Logger catches any panics and returns them as errors instead
type Logger struct{}

// Name of the module - fulfills Middleware interface
func (Logger) Name() string {
	return NameLogger
}

var _ stack.Middleware = Logger{}

// CheckTx logs time and result - fulfills Middlware interface
func (Logger) CheckTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Checker) (res sdk.CheckResult, err error) {
	start := time.Now()
	res, err = next.CheckTx(ctx, store, tx)
	delta := time.Now().Sub(start)
	// TODO: log some info on the tx itself?
	l := ctx.With("duration", micros(delta))
	if err == nil {
		l.Debug("CheckTx", "log", res.Log)
	} else {
		l.Info("CheckTx", "err", err)
	}
	return
}

// DeliverTx logs time and result - fulfills Middlware interface
func (Logger) DeliverTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Deliver) (res sdk.DeliverResult, err error) {
	start := time.Now()
	res, err = next.DeliverTx(ctx, store, tx)
	delta := time.Now().Sub(start)
	// TODO: log some info on the tx itself?
	l := ctx.With("duration", micros(delta))
	if err == nil {
		l.Info("DeliverTx", "log", res.Log)
	} else {
		l.Error("DeliverTx", "err", err)
	}
	return
}

// InitState logs time and result - fulfills Middlware interface
func (Logger) InitState(l log.Logger, store state.SimpleDB, module, key, value string, next sdk.InitStater) (string, error) {
	start := time.Now()
	res, err := next.InitState(l, store, module, key, value)
	delta := time.Now().Sub(start)
	// TODO: log the value being set also?
	l = l.With("duration", micros(delta)).With("mod", module).With("key", key)
	if err == nil {
		l.Info("InitState", "log", res)
	} else {
		l.Error("InitState", "err", err)
	}
	return res, err
}

// InitValidate logs time and result - fulfills Middlware interface
func (Logger) InitValidate(l log.Logger, store state.SimpleDB, vals []*abci.Validator, next sdk.InitValidater) {
	start := time.Now()
	next.InitValidate(l, store, vals)
	delta := time.Now().Sub(start)
	l = l.With("duration", micros(delta))
	l.Info("InitValidate")
}

// micros returns how many microseconds passed in a call
func micros(d time.Duration) int {
	return int(d.Seconds() * 1000000)
}
