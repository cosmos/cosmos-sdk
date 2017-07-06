package fee

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/types"
)

// NameFee - namespace for the fee module
const NameFee = "fee"

// AccountChecker - interface used by SimpleFeeHandler
type AccountChecker interface {
	// Get amount checks the current amount
	GetAmount(store state.KVStore, addr basecoin.Actor) (types.Coins, error)

	// ChangeAmount modifies the balance by the given amount and returns the new balance
	// always returns an error if leading to negative balance
	ChangeAmount(store state.KVStore, addr basecoin.Actor, coins types.Coins) (types.Coins, error)
}

// SimpleFeeHandler - checker object for fee checking
type SimpleFeeHandler struct {
	AccountChecker
	MinFee types.Coins
	stack.PassOption
}

// Name - return the namespace for the fee module
func (SimpleFeeHandler) Name() string {
	return NameFee
}

var _ stack.Middleware = SimpleFeeHandler{}

// Yes, I know refactor a bit... really too late already

// CheckTx - check the transaction
func (h SimpleFeeHandler) CheckTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Checker) (res basecoin.Result, err error) {
	feeTx, ok := tx.Unwrap().(*Fee)
	if !ok {
		return res, errors.ErrInvalidFormat(tx)
	}

	fees := types.Coins{feeTx.Fee}
	if !fees.IsGTE(h.MinFee) {
		return res, ErrInsufficientFees()
	}

	if !ctx.HasPermission(feeTx.Payer) {
		return res, errors.ErrUnauthorized()
	}

	_, err = h.ChangeAmount(store, feeTx.Payer, fees.Negative())
	if err != nil {
		return res, err
	}

	return basecoin.Result{Log: "Valid tx"}, nil
}

// DeliverTx - send the fee handler transaction
func (h SimpleFeeHandler) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.Result, err error) {
	feeTx, ok := tx.Unwrap().(*Fee)
	if !ok {
		return res, errors.ErrInvalidFormat(tx)
	}

	fees := types.Coins{feeTx.Fee}
	if !fees.IsGTE(h.MinFee) {
		return res, ErrInsufficientFees()
	}

	if !ctx.HasPermission(feeTx.Payer) {
		return res, errors.ErrUnauthorized()
	}

	_, err = h.ChangeAmount(store, feeTx.Payer, fees.Negative())
	if err != nil {
		return res, err
	}

	return next.DeliverTx(ctx, store, feeTx.Next())
}
