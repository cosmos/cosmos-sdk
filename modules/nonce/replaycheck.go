package nonce

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
)

//nolint
const (
	NameNonce = "nonce"
	CostNonce = 10
)

// ReplayCheck uses the sequence to check for replay attacks
type ReplayCheck struct {
	stack.PassInitState
	stack.PassInitValidate
}

// Name of the module - fulfills Middleware interface
func (ReplayCheck) Name() string {
	return NameNonce
}

var _ stack.Middleware = ReplayCheck{}

// CheckTx verifies tx is not being replayed - fulfills Middlware interface
func (r ReplayCheck) CheckTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, next sdk.Checker) (res sdk.CheckResult, err error) {

	stx, err := r.checkIncrementNonceTx(ctx, store, tx)
	if err != nil {
		return res, err
	}

	res, err = next.CheckTx(ctx, store, stx)
	res.GasAllocated += CostNonce
	return res, err
}

// DeliverTx verifies tx is not being replayed - fulfills Middlware interface
// NOTE It is okay to modify the sequence before running the wrapped TX because if the
// wrapped Tx fails, the state changes are not applied
func (r ReplayCheck) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, next sdk.Deliver) (res sdk.DeliverResult, err error) {

	stx, err := r.checkIncrementNonceTx(ctx, store, tx)
	if err != nil {
		return res, err
	}

	return next.DeliverTx(ctx, store, stx)
}

// checkNonceTx varifies the nonce sequence, an increment sequence number
func (r ReplayCheck) checkIncrementNonceTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx) (sdk.Tx, error) {

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
