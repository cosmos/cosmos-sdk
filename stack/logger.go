package stack

import (
	"time"

	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/types"
)

const (
	NameLogger = "lggr"
)

// Logger catches any panics and returns them as errors instead
type Logger struct{}

func (_ Logger) Name() string {
	return NameLogger
}

var _ Middleware = Logger{}

func (_ Logger) CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Checker) (res basecoin.Result, err error) {
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

func (_ Logger) DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.Result, err error) {
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

func (_ Logger) SetOption(l log.Logger, store types.KVStore, key, value string, next basecoin.SetOptioner) (string, error) {
	start := time.Now()
	res, err := next.SetOption(l, store, key, value)
	delta := time.Now().Sub(start)
	// TODO: log the value being set also?
	l = l.With("duration", micros(delta)).With("key", key)
	if err == nil {
		l.Info("SetOption", "log", res)
	} else {
		l.Error("SetOption", "err", err)
	}
	return res, err
}

// micros returns how many microseconds passed in a call
func micros(d time.Duration) int {
	return int(d.Seconds() * 1000000)
}
