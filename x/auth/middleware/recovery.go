package middleware

import (
	"context"
	"runtime/debug"

	abci "github.com/tendermint/tendermint/abci/types"

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
func (txh recoveryTxHandler) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (res abci.ResponseCheckTx, err error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Panic recovery.
	defer func() {
		if r := recover(); r != nil {
			err = handleRecovery(r, sdkCtx)
		}
	}()

	return txh.next.CheckTx(ctx, tx, req)
}

// DeliverTx implements tx.Handler.DeliverTx method.
func (txh recoveryTxHandler) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (res abci.ResponseDeliverTx, err error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// only run the tx if there is block gas remaining
	if sdkCtx.BlockGasMeter().IsOutOfGas() {
		err = sdkerrors.Wrap(sdkerrors.ErrOutOfGas, "no block gas left to run tx")
		return
	}

	startingGas := sdkCtx.BlockGasMeter().GasConsumed()

	// Panic recovery.
	defer func() {
		if r := recover(); r != nil {
			err = handleRecovery(r, sdkCtx)
		}
	}()

	// If BlockGasMeter() panics it will be caught by the above recover and will
	// return an error - in any case BlockGasMeter will consume gas past the limit.
	//
	// NOTE: This must exist in a separate defer function for the above recovery
	// to recover from this one.
	defer func() {
		sdkCtx.BlockGasMeter().ConsumeGas(
			sdkCtx.GasMeter().GasConsumedToLimit(), "block gas meter",
		)

		if sdkCtx.BlockGasMeter().GasConsumed() < startingGas {
			panic(sdk.ErrorGasOverflow{Descriptor: "tx gas summation"})
		}
	}()

	return txh.next.DeliverTx(ctx, tx, req)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (txh recoveryTxHandler) SimulateTx(ctx context.Context, sdkTx sdk.Tx, req tx.RequestSimulateTx) (res tx.ResponseSimulateTx, err error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Panic recovery.
	defer func() {
		if r := recover(); r != nil {
			err = handleRecovery(r, sdkCtx)
		}
	}()

	return txh.next.SimulateTx(ctx, sdkTx, req)
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
