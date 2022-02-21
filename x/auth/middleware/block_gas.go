package middleware

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

type consumeBlockGasHandler struct {
	next tx.Handler
}

// ConsumeBlockGasMiddleware check and consume block gas meter.
func ConsumeBlockGasMiddleware(txh tx.Handler) tx.Handler {
	return consumeBlockGasHandler{next: txh}
}

var _ tx.Handler = consumeBlockGasHandler{}

// CheckTx implements tx.Handler.CheckTx method.
func (cbgh consumeBlockGasHandler) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (res tx.Response, resCheckTx tx.ResponseCheckTx, err error) {
	return cbgh.next.CheckTx(ctx, req, checkReq)
}

// DeliverTx implements tx.Handler.DeliverTx method.
// Consume block gas meter, panic when block gas meter exceeded,
// the panic should be caught by `RecoveryTxMiddleware`.
func (cbgh consumeBlockGasHandler) DeliverTx(ctx context.Context, req tx.Request) (res tx.Response, err error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// only run the tx if there is block gas remaining
	if sdkCtx.BlockGasMeter().IsOutOfGas() {
		err = sdkerrors.Wrap(sdkerrors.ErrOutOfGas, "no block gas left to run tx")
		return
	}

	// If BlockGasMeter() panics it will be caught by the `RecoveryTxMiddleware` and will
	// return an error - in any case BlockGasMeter will consume gas past the limit.
	defer func() {
		sdkCtx.BlockGasMeter().ConsumeGas(
			sdkCtx.GasMeter().GasConsumedToLimit(), "block gas meter",
		)

	}()

	return cbgh.next.DeliverTx(ctx, req)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (cbgh consumeBlockGasHandler) SimulateTx(ctx context.Context, req tx.Request) (res tx.Response, err error) {
	return cbgh.next.SimulateTx(ctx, req)
}
