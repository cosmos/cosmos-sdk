package base

import (
	"fmt"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

//nolint
const (
	NameCheckpoint = "check"
)

// Checkpoint isolates all data store below this
type Checkpoint struct {
	stack.PassOption
}

// Name of the module - fulfills Middleware interface
func (Checkpoint) Name() string {
	return NameCheckpoint
}

var _ stack.Middleware = Chain{}

// CheckTx reverts all data changes if there was an error
func (c Checkpoint) CheckTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Checker) (res basecoin.Result, err error) {
	ps := state.NewKVCache(store)
	res, err = next.CheckTx(ctx, ps, tx)
	if err == nil {
		ps.Sync()
	}
	return res, err
}

// DeliverTx reverts all data changes if there was an error
func (c Checkpoint) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.Result, err error) {
	ps := state.NewKVCache(store)
	res, err = next.DeliverTx(ctx, ps, tx)
	if err == nil {
		fmt.Println("sync")
		ps.Sync()
	} else {
		fmt.Println("reject")
	}
	return res, err
}
