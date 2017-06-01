package basecoin

import (
	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire/data"

	"github.com/tendermint/basecoin/types"
)

// Handler is anything that processes a transaction
type Handler interface {
	CheckTx(ctx Context, store types.KVStore, tx Tx) (Result, error)
	DeliverTx(ctx Context, store types.KVStore, tx Tx) (Result, error)

	// TODO: flesh these out as well
	// SetOption(store types.KVStore, key, value string) (log string)
	// InitChain(store types.KVStore, vals []*abci.Validator)
	// BeginBlock(store types.KVStore, hash []byte, header *abci.Header)
	// EndBlock(store types.KVStore, height uint64) abci.ResponseEndBlock
}

// TODO: Context is a place-holder, soon we add some request data here from the
// higher-levels (like tell an app who signed).
// Trust me, we will need it like CallContext now...
type Context struct {
	sigs []crypto.PubKey
}

// TOTALLY insecure.  will redo later, but you get the point
func (c Context) AddSigners(keys ...crypto.PubKey) Context {
	return Context{
		sigs: append(c.sigs, keys...),
	}
}

func (c Context) GetSigners() []crypto.PubKey {
	return c.sigs
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
