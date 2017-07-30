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

// Bank is a default location for the fees, but pass anything into
// the middleware constructor
var Bank = basecoin.Actor{App: NameFee, Address: []byte("bank")}

// SimpleFeeMiddleware - middleware for fee checking, constant amount
// It used modules.coin to move the money
type SimpleFeeMiddleware struct {
	// the fee must be the same denomination and >= this amount
	// if the amount is 0, then the fee tx wrapper is optional
	MinFee coin.Coin
	// all fees go here, which could be a dump (Bank) or something reachable
	// by other app logic
	Collector basecoin.Actor
	stack.PassInitState
	stack.PassInitValidate
}

var _ stack.Middleware = SimpleFeeMiddleware{}

// NewSimpleFeeMiddleware returns a fee handler with a fixed minimum fee.
//
// If minFee is 0, then the FeeTx is optional
func NewSimpleFeeMiddleware(minFee coin.Coin, collector basecoin.Actor) SimpleFeeMiddleware {
	return SimpleFeeMiddleware{
		MinFee:    minFee,
		Collector: collector,
	}
}

// Name - return the namespace for the fee module
func (SimpleFeeMiddleware) Name() string {
	return NameFee
}

// CheckTx - check the transaction
func (h SimpleFeeMiddleware) CheckTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx, next basecoin.Checker) (res basecoin.CheckResult, err error) {
	fee, err := h.verifyFee(ctx, tx)
	if IsSkipFeesErr(err) {
		return next.CheckTx(ctx, store, tx)
	}
	if err != nil {
		return res, err
	}

	if !fee.Fee.IsZero() { // now, try to make a IPC call to coins...
		send := coin.NewSendOneTx(fee.Payer, h.Collector, coin.Coins{fee.Fee})
		_, err = next.CheckTx(ctx, store, send)
		if err != nil {
			return res, err
		}
	}

	return next.CheckTx(ctx, store, fee.Tx)
}

// DeliverTx - send the fee handler transaction
func (h SimpleFeeMiddleware) DeliverTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.DeliverResult, err error) {
	fee, err := h.verifyFee(ctx, tx)
	if IsSkipFeesErr(err) {
		return next.DeliverTx(ctx, store, tx)
	}
	if err != nil {
		return res, err
	}

	if !fee.Fee.IsZero() { // now, try to make a IPC call to coins...
		send := coin.NewSendOneTx(fee.Payer, h.Collector, coin.Coins{fee.Fee})
		_, err = next.DeliverTx(ctx, store, send)
		if err != nil {
			return res, err
		}
	}
	return next.DeliverTx(ctx, store, fee.Tx)
}

func (h SimpleFeeMiddleware) verifyFee(ctx basecoin.Context, tx basecoin.Tx) (Fee, error) {
	feeTx, ok := tx.Unwrap().(Fee)
	if !ok {
		// the fee wrapper is not required if there is no minimum
		if h.MinFee.IsZero() {
			return feeTx, ErrSkipFees()
		}
		return feeTx, errors.ErrInvalidFormat(TypeFees, tx)
	}

	// see if it is the proper denom and big enough
	fee := feeTx.Fee
	if fee.Denom != h.MinFee.Denom {
		return feeTx, ErrWrongFeeDenom(h.MinFee.Denom)
	}
	if !fee.IsGTE(h.MinFee) {
		return feeTx, ErrInsufficientFees()
	}
	return feeTx, nil
}
