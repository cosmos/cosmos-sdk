package middleware

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

type errorTxHandler struct {
	inner tx.TxHandler
	debug bool
}

// NewErrorTxMiddleware is a middleware that converts an error from inner
// middlewares into a abci.Response{Check,Deliver}Tx. It should generally act
// as the outermost middleware.
func NewErrorTxMiddleware(debug bool) tx.TxMiddleware {
	return func(txh tx.TxHandler) tx.TxHandler {
		return errorTxHandler{inner: txh, debug: debug}
	}

}

var _ tx.TxHandler = errorTxHandler{}

// CheckTx implements TxHandler.CheckTx.
func (txh errorTxHandler) CheckTx(ctx sdk.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	res, err := txh.inner.CheckTx(ctx, tx, req)
	if err != nil {
		gInfo := sdk.GasInfo{GasUsed: ctx.BlockGasMeter().GasConsumed()}

		return sdkerrors.ResponseCheckTx(err, gInfo.GasWanted, gInfo.GasUsed, txh.debug), nil
	}

	return res, nil
}

// DeliverTx implements TxHandler.DeliverTx.
func (txh errorTxHandler) DeliverTx(ctx sdk.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	res, err := txh.inner.DeliverTx(ctx, tx, req)
	if err != nil {
		gInfo := sdk.GasInfo{GasUsed: ctx.BlockGasMeter().GasConsumed()}

		return sdkerrors.ResponseDeliverTx(err, gInfo.GasWanted, gInfo.GasUsed, txh.debug), nil
	}

	return res, nil
}
