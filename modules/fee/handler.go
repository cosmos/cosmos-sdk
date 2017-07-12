package fee

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

// NameFee - namespace for the fee module
const NameFee = "fee"

// SimpleFeeMiddleware - middleware for fee checking, constant amount
// It used modules.coin to move the money
type SimpleFeeMiddleware struct {
	MinFee    coin.Coin //
	Collector basecoin.Actor
	stack.PassOption
}

var _ stack.Middleware = SimpleFeeMiddleware{}

// NewSimpleFeeMiddleware returns a fee handler with a fixed minimum fee.
//
// If minFee is 0, then the FeeTx is optional
func NewSimpleFeeMiddleware(minFee coin.Coin) SimpleFeeMiddleware {
	return SimpleFeeMiddleware{
		MinFee: minFee,
	}
}

// Name - return the namespace for the fee module
func (SimpleFeeMiddleware) Name() string {
	return NameFee
}

// CheckTx - check the transaction
func (h SimpleFeeMiddleware) CheckTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Checker) (res basecoin.Result, err error) {
	return h.doTx(ctx, store, tx, next.CheckTx)
}

// DeliverTx - send the fee handler transaction
func (h SimpleFeeMiddleware) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.Result, err error) {
	return h.doTx(ctx, store, tx, next.DeliverTx)
}

func (h SimpleFeeMiddleware) doTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.CheckerFunc) (res basecoin.Result, err error) {
	feeTx, ok := tx.Unwrap().(Fee)
	if !ok {
		// the fee wrapper is not required if there is no minimum
		if h.MinFee.IsZero() {
			return next(ctx, store, tx)
		}
		return res, errors.ErrInvalidFormat(TypeFees, tx)
	}

	// see if it is big enough...
	fee := feeTx.Fee
	if !fee.IsGTE(h.MinFee) {
		return res, ErrInsufficientFees()
	}

	// now, try to make a IPC call to coins...
	send := coin.NewSendOneTx(feeTx.Payer, h.Collector, coin.Coins{fee})
	_, err = next(ctx, store, send)
	if err != nil {
		return res, err
	}

	return next(ctx, store, feeTx.Tx)
}
