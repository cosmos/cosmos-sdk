package middleware

import (
	"fmt"
	"runtime/debug"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

type panicTxHandler struct {
	inner tx.TxHandler
	debug bool
}

func NewPanicTxMiddleware(debug bool) tx.TxMiddleware {
	return func(txh tx.TxHandler) tx.TxHandler {
		return panicTxHandler{inner: txh, debug: debug}
	}

}

var _ tx.TxHandler = panicTxHandler{}

func (txh panicTxHandler) CheckTx(ctx sdk.Context, tx sdk.Tx, req abci.RequestCheckTx) (res abci.ResponseCheckTx, err error) {
	// Panic recovery.
	defer func() {
		// GasMeter expected to be set in AnteHandler
		gasWanted := ctx.GasMeter().Limit()

		if r := recover(); r != nil {
			recoveryMW := newOutOfGasRecoveryMiddleware(gasWanted, ctx, newDefaultRecoveryMiddleware())
			err = processRecovery(r, recoveryMW)
		}

		gInfo := sdk.GasInfo{GasWanted: gasWanted, GasUsed: ctx.GasMeter().GasConsumed()}
		res, err = sdkerrors.ResponseCheckTx(err, gInfo.GasWanted, gInfo.GasUsed, txh.debug), nil
	}()

	return txh.inner.CheckTx(ctx, tx, req)
}

func (txh panicTxHandler) DeliverTx(ctx sdk.Context, tx sdk.Tx, req abci.RequestDeliverTx) (res abci.ResponseDeliverTx, err error) {
	// only run the tx if there is block gas remaining
	if ctx.BlockGasMeter().IsOutOfGas() {
		gInfo := sdk.GasInfo{GasUsed: ctx.BlockGasMeter().GasConsumed()}
		return sdkerrors.ResponseDeliverTx(sdkerrors.Wrap(sdkerrors.ErrOutOfGas, "no block gas left to run tx"), gInfo.GasWanted, gInfo.GasUsed, txh.debug), nil
	}

	startingGas := ctx.BlockGasMeter().GasConsumed()

	// Panic recovery.
	defer func() {
		// GasMeter expected to be set in AnteHandler
		gasWanted := ctx.GasMeter().Limit()

		if r := recover(); r != nil {
			recoveryMW := newOutOfGasRecoveryMiddleware(gasWanted, ctx, newDefaultRecoveryMiddleware())
			err = processRecovery(r, recoveryMW)
		}

		gInfo := sdk.GasInfo{GasWanted: gasWanted, GasUsed: ctx.GasMeter().GasConsumed()}
		res, err = sdkerrors.ResponseDeliverTx(err, gInfo.GasWanted, gInfo.GasUsed, txh.debug), nil
	}()

	// If BlockGasMeter() panics it will be caught by the above recover and will
	// return an error - in any case BlockGasMeter will consume gas past the limit.
	//
	// NOTE: This must exist in a separate defer function for the above recovery
	// to recover from this one.
	defer func() {
		ctx.BlockGasMeter().ConsumeGas(
			ctx.GasMeter().GasConsumedToLimit(), "block gas meter",
		)

		if ctx.BlockGasMeter().GasConsumed() < startingGas {
			panic(sdk.ErrorGasOverflow{Descriptor: "tx gas summation"})
		}
	}()

	return txh.inner.DeliverTx(ctx, tx, req)
}

// RecoveryHandler handles recovery() object.
// Return a non-nil error if recoveryObj was processed.
// Return nil if recoveryObj was not processed.
type recoveryHandler func(recoveryObj interface{}) error

// recoveryMiddleware is wrapper for RecoveryHandler to create chained recovery handling.
// returns (recoveryMiddleware, nil) if recoveryObj was not processed and should be passed to the next middleware in chain.
// returns (nil, error) if recoveryObj was processed and middleware chain processing should be stopped.
type recoveryMiddleware func(recoveryObj interface{}) (recoveryMiddleware, error)

// processRecovery processes recoveryMiddleware chain for recovery() object.
// Chain processing stops on non-nil error or when chain is processed.
func processRecovery(recoveryObj interface{}, middleware recoveryMiddleware) error {
	if middleware == nil {
		return nil
	}

	next, err := middleware(recoveryObj)
	if err != nil {
		return err
	}

	return processRecovery(recoveryObj, next)
}

// newRecoveryMiddleware creates a RecoveryHandler middleware.
func newRecoveryMiddleware(handler recoveryHandler, next recoveryMiddleware) recoveryMiddleware {
	return func(recoveryObj interface{}) (recoveryMiddleware, error) {
		if err := handler(recoveryObj); err != nil {
			return nil, err
		}

		return next, nil
	}
}

// newOutOfGasRecoveryMiddleware creates a standard OutOfGas recovery middleware for app.runTx method.
func newOutOfGasRecoveryMiddleware(gasWanted uint64, ctx sdk.Context, next recoveryMiddleware) recoveryMiddleware {
	handler := func(recoveryObj interface{}) error {
		err, ok := recoveryObj.(sdk.ErrorOutOfGas)
		if !ok {
			return nil
		}

		return sdkerrors.Wrap(
			sdkerrors.ErrOutOfGas, fmt.Sprintf(
				"out of gas in location: %v; gasWanted: %d, gasUsed: %d",
				err.Descriptor, gasWanted, ctx.GasMeter().GasConsumed(),
			),
		)
	}

	return newRecoveryMiddleware(handler, next)
}

// newDefaultRecoveryMiddleware creates a default (last in chain) recovery middleware for app.runTx method.
func newDefaultRecoveryMiddleware() recoveryMiddleware {
	handler := func(recoveryObj interface{}) error {
		return sdkerrors.Wrap(
			sdkerrors.ErrPanic, fmt.Sprintf(
				"recovered: %v\nstack:\n%v", recoveryObj, string(debug.Stack()),
			),
		)
	}

	return newRecoveryMiddleware(handler, nil)
}
