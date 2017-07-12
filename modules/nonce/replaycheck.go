package nonce

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

//nolint
const (
	NameNonce = "nonce"
)

// ReplayCheck parses out go-crypto signatures and adds permissions to the
// context for use inside the application
type ReplayCheck struct {
	stack.PassOption
}

// Name of the module - fulfills Middleware interface
func (ReplayCheck) Name() string {
	return NameNonce
}

var _ stack.Middleware = ReplayCheck{}

// CheckTx verifies the signatures are correct - fulfills Middlware interface
func (ReplayCheck) CheckTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Checker) (res basecoin.Result, err error) {
	sigs, tnext, err := getSigners(tx)
	if err != nil {
		return res, err
	}
	ctx2 := addSigners(ctx, sigs)
	return next.CheckTx(ctx2, store, tnext)
}

// DeliverTx verifies the signatures are correct - fulfills Middlware interface
func (ReplayCheck) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.Result, err error) {
	sigs, tnext, err := getSigners(tx)
	if err != nil {
		return res, err
	}
	ctx2 := addSigners(ctx, sigs)
	return next.DeliverTx(ctx2, store, tnext)
}
