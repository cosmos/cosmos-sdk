package stack

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/types"
)

type CheckerMiddle interface {
	CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Checker) (basecoin.Result, error)
}

type DeliverMiddle interface {
	DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Deliver) (basecoin.Result, error)
}

// Middleware is anything that wraps another handler to enhance functionality.
//
// You can use utilities in handlers to construct them, the interfaces
// are exposed in the top-level package to avoid import loops.
type Middleware interface {
	CheckerMiddle
	DeliverMiddle
	basecoin.Named
}
