package stack

import (
	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/types"
)

// Middleware is anything that wraps another handler to enhance functionality.
//
// You can use utilities in handlers to construct them, the interfaces
// are exposed in the top-level package to avoid import loops.
type Middleware interface {
	CheckerMiddle
	DeliverMiddle
	SetOptionMiddle
	basecoin.Named
}

type CheckerMiddle interface {
	CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Checker) (basecoin.Result, error)
}

type CheckerMiddleFunc func(basecoin.Context, types.KVStore, basecoin.Tx, basecoin.Checker) (basecoin.Result, error)

func (c CheckerMiddleFunc) CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Checker) (basecoin.Result, error) {
	return c(ctx, store, tx, next)
}

type DeliverMiddle interface {
	DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Deliver) (basecoin.Result, error)
}

type DeliverMiddleFunc func(basecoin.Context, types.KVStore, basecoin.Tx, basecoin.Deliver) (basecoin.Result, error)

func (d DeliverMiddleFunc) DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Deliver) (basecoin.Result, error) {
	return d(ctx, store, tx, next)
}

type SetOptionMiddle interface {
	SetOption(l log.Logger, store types.KVStore, key, value string, next basecoin.SetOptioner) (string, error)
}

type SetOptionMiddleFunc func(log.Logger, types.KVStore, string, string, basecoin.SetOptioner) (string, error)

func (c SetOptionMiddleFunc) SetOption(l log.Logger, store types.KVStore, key, value string, next basecoin.SetOptioner) (string, error) {
	return c(l, store, key, value, next)
}

// holders
type PassCheck struct{}

func (_ PassCheck) CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Checker) (basecoin.Result, error) {
	return next.CheckTx(ctx, store, tx)
}

type PassDeliver struct{}

func (_ PassDeliver) DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Deliver) (basecoin.Result, error) {
	return next.DeliverTx(ctx, store, tx)
}

type PassOption struct{}

func (_ PassOption) SetOption(l log.Logger, store types.KVStore, key, value string, next basecoin.SetOptioner) (string, error) {
	return next.SetOption(l, store, key, value)
}
