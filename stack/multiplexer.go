package stack

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/txs"
	"github.com/tendermint/basecoin/types"
)

const (
	NameMultiplexer = "mplx"
)

type Multiplexer struct{}

func (_ Multiplexer) Name() string {
	return NameMultiplexer
}

var _ Middleware = Multiplexer{}

func (_ Multiplexer) CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Checker) (res basecoin.Result, err error) {
	if mtx, ok := tx.Unwrap().(*txs.MultiTx); ok {
		return runAll(ctx, store, mtx.Txs, next.CheckTx)
	}
	return next.CheckTx(ctx, store, tx)
}

func (_ Multiplexer) DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.Result, err error) {
	if mtx, ok := tx.Unwrap().(*txs.MultiTx); ok {
		return runAll(ctx, store, mtx.Txs, next.DeliverTx)
	}
	return next.DeliverTx(ctx, store, tx)
}

func runAll(ctx basecoin.Context, store types.KVStore, txs []basecoin.Tx, next basecoin.CheckerFunc) (res basecoin.Result, err error) {
	// store all results, unless anything errors
	rs := make([]basecoin.Result, len(txs))
	for i, stx := range txs {
		rs[i], err = next(ctx, store, stx)
		if err != nil {
			return
		}
	}
	// now combine the results into one...
	return combine(rs), nil
}

func combine(res []basecoin.Result) basecoin.Result {
	// TODO: how to combine???
	return res[0]
}
