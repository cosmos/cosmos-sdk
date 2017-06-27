package handlers

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/txs"
	"github.com/tendermint/basecoin/types"
)

const (
	NameFee = "fee"
)

type AccountChecker interface {
	// Get amount checks the current amount
	GetAmount(store types.KVStore, addr []byte) (types.Coins, error)

	// ChangeAmount modifies the balance by the given amount and returns the new balance
	// always returns an error if leading to negative balance
	ChangeAmount(store types.KVStore, addr []byte, coins types.Coins) (types.Coins, error)
}

type SimpleFeeHandler struct {
	AccountChecker
	MinFee types.Coins
}

func (_ SimpleFeeHandler) Name() string {
	return NameFee
}

var _ basecoin.Middleware = SimpleFeeHandler{}

// Yes, I know refactor a bit... really too late already

func (h SimpleFeeHandler) CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Checker) (res basecoin.Result, err error) {
	feeTx, ok := tx.Unwrap().(*txs.Fee)
	if !ok {
		return res, errors.InvalidFormat()
	}

	fees := types.Coins{feeTx.Fee}
	if !fees.IsGTE(h.MinFee) {
		return res, errors.InsufficientFees()
	}

	if !ctx.HasPermission(NameSigs, feeTx.Payer) {
		return res, errors.Unauthorized()
	}

	_, err = h.ChangeAmount(store, feeTx.Payer, fees.Negative())
	if err != nil {
		return res, err
	}

	return basecoin.Result{Log: "Valid tx"}, nil
}

func (h SimpleFeeHandler) DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.Result, err error) {
	feeTx, ok := tx.Unwrap().(*txs.Fee)
	if !ok {
		return res, errors.InvalidFormat()
	}

	fees := types.Coins{feeTx.Fee}
	if !fees.IsGTE(h.MinFee) {
		return res, errors.InsufficientFees()
	}

	if !ctx.HasPermission(NameSigs, feeTx.Payer) {
		return res, errors.Unauthorized()
	}

	_, err = h.ChangeAmount(store, feeTx.Payer, fees.Negative())
	if err != nil {
		return res, err
	}

	return next.DeliverTx(ctx, store, feeTx.Next())
}
