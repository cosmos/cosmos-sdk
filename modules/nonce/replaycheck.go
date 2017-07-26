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

// ReplayCheck uses the sequence to check for replay attacks
type ReplayCheck struct {
	stack.PassOption
}

// Name of the module - fulfills Middleware interface
func (ReplayCheck) Name() string {
	return NameNonce
}

var _ stack.Middleware = ReplayCheck{}

// CheckTx verifies tx is not being replayed - fulfills Middlware interface
func (r ReplayCheck) CheckTx(ctx basecoin.Context, store state.SimpleDB,
	tx basecoin.Tx, next basecoin.Checker) (res basecoin.Result, err error) {

	stx, err := r.checkIncrementNonceTx(ctx, store, tx)
	if err != nil {
		return res, err
	}

	return next.CheckTx(ctx, store, stx)
}

// DeliverTx verifies tx is not being replayed - fulfills Middlware interface
// NOTE It is okay to modify the sequence before running the wrapped TX because if the
// wrapped Tx fails, the state changes are not applied
func (r ReplayCheck) DeliverTx(ctx basecoin.Context, store state.SimpleDB,
	tx basecoin.Tx, next basecoin.Deliver) (res basecoin.Result, err error) {

	stx, err := r.checkIncrementNonceTx(ctx, store, tx)
	if err != nil {
		return res, err
	}

	return next.DeliverTx(ctx, store, stx)
}

// checkNonceTx varifies the nonce sequence, an increment sequence number
func (r ReplayCheck) checkIncrementNonceTx(ctx basecoin.Context, store state.SimpleDB,
	tx basecoin.Tx) (basecoin.Tx, error) {

	// make sure it is a the nonce Tx (Tx from this package)
	nonceTx, ok := tx.Unwrap().(Tx)
	if !ok {
		return tx, ErrNoNonce()
	}

	err := nonceTx.ValidateBasic()
	if err != nil {
		return tx, err
	}

	// check the nonce sequence number
	err = nonceTx.CheckIncrementSeq(ctx, store)
	if err != nil {
		return tx, err
	}
	return nonceTx.Tx, nil
}
