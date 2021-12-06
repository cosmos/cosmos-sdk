package middleware

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

type tipsTxHandler struct {
	next       tx.Handler
	bankKeeper types.BankKeeper
}

// NewTipMiddleware returns a new middleware for handling transactions with
// tips.
func NewTipMiddleware(bankKeeper types.BankKeeper) tx.Middleware {
	return func(txh tx.Handler) tx.Handler {
		return tipsTxHandler{txh, bankKeeper}
	}
}

var _ tx.Handler = tipsTxHandler{}

// CheckTx implements tx.Handler.CheckTx.
func (txh tipsTxHandler) CheckTx(ctx context.Context, req tx.Request, checkTx tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	res, resCheckTx, err := txh.next.CheckTx(ctx, req, checkTx)
	res, err = txh.transferTip(ctx, req, res, err)

	return res, resCheckTx, err
}

// DeliverTx implements tx.Handler.DeliverTx.
func (txh tipsTxHandler) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	res, err := txh.next.DeliverTx(ctx, req)

	return txh.transferTip(ctx, req, res, err)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (txh tipsTxHandler) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	res, err := txh.next.SimulateTx(ctx, req)

	return txh.transferTip(ctx, req, res, err)
}

// transferTip transfers the tip from the tipper to the fee payer.
func (txh tipsTxHandler) transferTip(ctx context.Context, req tx.Request, res tx.Response, err error) (tx.Response, error) {
	tipTx, ok := req.Tx.(tx.TipTx)

	// No-op if the tx doesn't have tips.
	if !ok || tipTx.GetTip() == nil {
		return res, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	tipper, err := sdk.AccAddressFromBech32(tipTx.GetTip().Tipper)
	if err != nil {
		return tx.Response{}, err
	}

	err = txh.bankKeeper.SendCoins(sdkCtx, tipper, tipTx.FeePayer(), tipTx.GetTip().Amount)
	if err != nil {
		return tx.Response{}, err
	}

	return res, nil
}
