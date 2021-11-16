package middleware

import (
	"context"

	abci "github.com/tendermint/tendermint/abci/types"

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
func (txh tipsTxHandler) CheckTx(ctx context.Context, sdkTx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	res, err := txh.next.CheckTx(ctx, sdkTx, req)
	if err != nil {
		return abci.ResponseCheckTx{}, err
	}

	tipTx, ok := sdkTx.(tx.TipTx)
	if !ok || tipTx.GetTip() == nil {
		return res, err
	}

	if err := txh.transferTip(ctx, tipTx); err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return res, err
}

// DeliverTx implements tx.Handler.DeliverTx.
func (txh tipsTxHandler) DeliverTx(ctx context.Context, sdkTx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	res, err := txh.next.DeliverTx(ctx, sdkTx, req)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	tipTx, ok := sdkTx.(tx.TipTx)
	if !ok || tipTx.GetTip() == nil {
		return res, err
	}

	if err := txh.transferTip(ctx, tipTx); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return res, err
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (txh tipsTxHandler) SimulateTx(ctx context.Context, sdkTx sdk.Tx, req tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	res, err := txh.next.SimulateTx(ctx, sdkTx, req)
	if err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	tipTx, ok := sdkTx.(tx.TipTx)
	if !ok || tipTx.GetTip() == nil {
		return res, err
	}

	if err := txh.transferTip(ctx, tipTx); err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	return res, err
}

// transferTip transfers the tip from the tipper to the fee payer.
func (txh tipsTxHandler) transferTip(ctx context.Context, tipTx tx.TipTx) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	tipper, err := sdk.AccAddressFromBech32(tipTx.GetTip().Tipper)
	if err != nil {
		return err
	}

	return txh.bankKeeper.SendCoins(sdkCtx, tipper, tipTx.FeePayer(), tipTx.GetTip().Amount)
}
