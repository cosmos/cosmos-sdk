package middleware

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

type txDecoderTxHandler struct {
	next      tx.Handler
	txDecoder sdk.TxDecoder
}

// TxDecoderMiddleware
func NewTxDecoderMiddleware(txDecoder sdk.TxDecoder) tx.Middleware {
	return func(txh tx.Handler) tx.Handler {
		return txDecoderTxHandler{next: txh, txDecoder: txDecoder}
	}
}

var _ tx.Handler = gasTxHandler{}

// CheckTx implements tx.Handler.CheckTx.
func (h txDecoderTxHandler) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	newReq, err := h.populateReq(req)
	if err != nil {
		return tx.Response{}, tx.ResponseCheckTx{}, err
	}

	return h.next.CheckTx(ctx, newReq, checkReq)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (h txDecoderTxHandler) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	newReq, err := h.populateReq(req)
	if err != nil {
		return tx.Response{}, err
	}

	return h.next.DeliverTx(ctx, newReq)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (h txDecoderTxHandler) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	newReq, err := h.populateReq(req)
	if err != nil {
		return tx.Response{}, err
	}

	return h.next.SimulateTx(ctx, newReq)
}

// populateReq takes a tx.Request, and if its Tx field is not set, then
// decodes the TxBytes and populates the decoded Tx field.
func (h txDecoderTxHandler) populateReq(req tx.Request) (tx.Request, error) {
	if len(req.TxBytes) == 0 && req.Tx == nil {
		return tx.Request{}, sdkerrors.ErrInvalidRequest.Wrap("got empty tx request")
	}

	sdkTx := req.Tx
	var err error
	if len(req.TxBytes) != 0 {
		sdkTx, err = h.txDecoder(req.TxBytes)
		if err != nil {
			return tx.Request{}, err
		}
	}

	return tx.Request{Tx: sdkTx, TxBytes: req.TxBytes}, nil
}
