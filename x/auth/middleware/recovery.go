package middleware

import (
	"context"
	"runtime/debug"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

type recoveryTxHandler struct {
	next tx.Handler
}

// RecoveryTxMiddleware defines a middleware that catches all panics that
// happen in inner middlewares.
//
// Be careful, it won't catch any panics happening outside!
func RecoveryTxMiddleware(txh tx.Handler) tx.Handler {
	return recoveryTxHandler{next: txh}
}

var _ tx.Handler = recoveryTxHandler{}

// CheckTx implements tx.Handler.CheckTx method.
func (txh recoveryTxHandler) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (res tx.Response, resCheckTx tx.ResponseCheckTx, err error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Panic recovery.
	defer func() {
		if r := recover(); r != nil {
			err = handleRecovery(r, sdkCtx)
		}
	}()

	return txh.next.CheckTx(ctx, req, checkReq)
}

// DeliverTx implements tx.Handler.DeliverTx method.
func (txh recoveryTxHandler) DeliverTx(ctx context.Context, req tx.Request) (res tx.Response, err error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Panic recovery.
	defer func() {
		if r := recover(); r != nil {
			err = handleRecovery(r, sdkCtx)
		}
	}()

	return txh.next.DeliverTx(ctx, req)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (txh recoveryTxHandler) SimulateTx(ctx context.Context, req tx.Request) (res tx.Response, err error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Panic recovery.
	defer func() {
		if r := recover(); r != nil {
			err = handleRecovery(r, sdkCtx)
		}
	}()

	return txh.next.SimulateTx(ctx, req)
}

func handleRecovery(r interface{}, sdkCtx sdk.Context) error {
	switch r := r.(type) {
	case sdk.ErrorOutOfGas:
		return sdkerrors.Wrapf(sdkerrors.ErrOutOfGas,
			"out of gas in location: %v; gasWanted: %d, gasUsed: %d",
			r.Descriptor, sdkCtx.GasMeter().Limit(), sdkCtx.GasMeter().GasConsumed(),
		)

	default:
		return sdkerrors.ErrPanic.Wrapf(
			"recovered: %v\nstack:\n%v", r, string(debug.Stack()),
		)
	}
}
