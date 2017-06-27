package basecoin

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire/data"

	"github.com/tendermint/basecoin/types"
)

type Named interface {
	Name() string
}

type Checker interface {
	CheckTx(ctx Context, store types.KVStore, tx Tx) (Result, error)
}

type CheckerMiddle interface {
	CheckTx(ctx Context, store types.KVStore, tx Tx, next Checker) (Result, error)
}

// CheckerFunc (like http.HandlerFunc) is a shortcut for making wrapers
type CheckerFunc func(Context, types.KVStore, Tx) (Result, error)

func (c CheckerFunc) CheckTx(ctx Context, store types.KVStore, tx Tx) (Result, error) {
	return c(ctx, store, tx)
}

var _ Checker = CheckerFunc(nil)

type Deliver interface {
	DeliverTx(ctx Context, store types.KVStore, tx Tx) (Result, error)
}

type DeliverMiddle interface {
	DeliverTx(ctx Context, store types.KVStore, tx Tx, next Deliver) (Result, error)
}

// DeliverFunc (like http.HandlerFunc) is a shortcut for making wrapers
type DeliverFunc func(Context, types.KVStore, Tx) (Result, error)

func (c DeliverFunc) DeliverTx(ctx Context, store types.KVStore, tx Tx) (Result, error) {
	return c(ctx, store, tx)
}

var _ Deliver = DeliverFunc(nil)

// Handler is anything that processes a transaction
type Handler interface {
	Checker
	Deliver
	Named
	// TODO: flesh these out as well
	// SetOption(store types.KVStore, key, value string) (log string)
	// InitChain(store types.KVStore, vals []*abci.Validator)
	// BeginBlock(store types.KVStore, hash []byte, header *abci.Header)
	// EndBlock(store types.KVStore, height uint64) abci.ResponseEndBlock
}

// Middleware is anything that wraps another handler to enhance functionality.
//
// You can use utilities in handlers to construct them, the interfaces
// are exposed in the top-level package to avoid import loops.
type Middleware interface {
	CheckerMiddle
	DeliverMiddle
	Named
}

// Result captures any non-error abci result
// to make sure people use error for error cases
type Result struct {
	Data data.Bytes
	Log  string
}

func (r Result) ToABCI() abci.Result {
	return abci.Result{
		Data: r.Data,
		Log:  r.Log,
	}
}
