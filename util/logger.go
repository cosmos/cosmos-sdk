package util

import (
	"time"

	abci "github.com/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk"
)

// Logger writes out log messages on every request
type Logger struct{}

var _ sdk.Decorator = Logger{}

// CheckTx logs time and result - fulfills Middlware interface
func (Logger) CheckTx(ctx sdk.Context, store sdk.SimpleDB, tx interface{}, next sdk.Checker) (res sdk.CheckResult, err error) {
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
func (Logger) DeliverTx(ctx sdk.Context, store sdk.SimpleDB, tx interface{}, next sdk.Deliverer) (res sdk.DeliverResult, err error) {
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

// LogTicker wraps any ticker and records the time it takes.
// Pass in a name to be logged with this to separate out various
// tickers
func LogTicker(clock sdk.Ticker, name string) sdk.Ticker {
	res := func(ctx sdk.Context, s sdk.SimpleDB) ([]*abci.Validator, error) {
		start := time.Now()
		vals, err := clock.Tick(ctx, s)
		delta := time.Now().Sub(start)
		l := ctx.With("duration", micros(delta))
		if name != "" {
			l = l.With("name", name)
		}
		if err == nil {
			l.Info("Tock")
		} else {
			l.Error("Tock", "err", err)
		}
		return vals, err
	}
	return sdk.TickerFunc(res)
}

// micros returns how many microseconds passed in a call
func micros(d time.Duration) int {
	return int(d.Seconds() * 1000000)
}
