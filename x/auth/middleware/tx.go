package middleware

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

type txDecoderHandler struct {
	next      tx.Handler
	txDecoder sdk.TxDecoder
}

// NewTxDecoderMiddleware creates a new middleware that will decode tx bytes
// into a sdk.Tx. As input request, at least one of Tx or TxBytes must be set.
// If only TxBytes is set, then TxDecoderMiddleware will populate the Tx field.
// If only Tx is set, then TxBytes will be left empty, but some middlewares
// such as signature verification might fail.
func NewTxDecoderMiddleware(txDecoder sdk.TxDecoder) tx.Middleware {
	return func(txh tx.Handler) tx.Handler {
		return txDecoderHandler{next: txh, txDecoder: txDecoder}
	}
}

var _ tx.Handler = gasTxHandler{}

// CheckTx implements tx.Handler.CheckTx.
func (h txDecoderHandler) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	newReq, err := h.populateReq(req)
	if err != nil {
		return tx.Response{}, tx.ResponseCheckTx{}, err
	}

	return h.next.CheckTx(ctx, newReq, checkReq)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (h txDecoderHandler) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	newReq, err := h.populateReq(req)
	if err != nil {
		return tx.Response{}, err
	}

	return h.next.DeliverTx(ctx, newReq)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (h txDecoderHandler) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	newReq, err := h.populateReq(req)
	if err != nil {
		return tx.Response{}, err
	}

	return h.next.SimulateTx(ctx, newReq)
}

// populateReq takes a tx.Request, and if its Tx field is not set, then
// decodes the TxBytes and populates the decoded Tx field. It leaves
// req.TxBytes untouched.
func (h txDecoderHandler) populateReq(req tx.Request) (tx.Request, error) {
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
