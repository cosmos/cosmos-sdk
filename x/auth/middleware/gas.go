package middleware

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// GasTx defines a Tx with a GetGas() method which is needed to use gasTxHandler.
type GasTx interface {
	sdk.Tx
	GetGas() uint64
}

type gasTxHandler struct {
	next tx.Handler
}

// GasTxMiddleware defines a simple middleware that sets a new GasMeter on
// the sdk.Context, and sets the GasInfo on the result. It reads the tx.GetGas()
// by default, or sets to infinity in simulate mode.
func GasTxMiddleware(txh tx.Handler) tx.Handler {
	return gasTxHandler{next: txh}
}

var _ tx.Handler = gasTxHandler{}

// CheckTx implements tx.Handler.CheckTx.
func (txh gasTxHandler) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	sdkCtx, err := gasContext(sdk.UnwrapSDKContext(ctx), req.Tx, false)
	if err != nil {
		return tx.Response{}, tx.ResponseCheckTx{}, err
	}

	res, resCheckTx, err := txh.next.CheckTx(sdk.WrapSDKContext(sdkCtx), req, checkReq)

	return populateGas(res, sdkCtx), resCheckTx, err
}

// DeliverTx implements tx.Handler.DeliverTx.
func (txh gasTxHandler) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	sdkCtx, err := gasContext(sdk.UnwrapSDKContext(ctx), req.Tx, false)
	if err != nil {
		return tx.Response{}, err
	}

	res, err := txh.next.DeliverTx(sdk.WrapSDKContext(sdkCtx), req)

	return populateGas(res, sdkCtx), err
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (txh gasTxHandler) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	sdkCtx, err := gasContext(sdk.UnwrapSDKContext(ctx), req.Tx, true)
	if err != nil {
		return tx.Response{}, err
	}

	res, err := txh.next.SimulateTx(sdk.WrapSDKContext(sdkCtx), req)

	return populateGas(res, sdkCtx), err
}

// populateGas returns a new tx.Response with gas fields populated.
func populateGas(res tx.Response, sdkCtx sdk.Context) tx.Response {
	res.GasWanted = sdkCtx.GasMeter().Limit()
	res.GasUsed = sdkCtx.GasMeter().GasConsumed()

	return res
}

// gasContext returns a new context with a gas meter set from a given context.
func gasContext(ctx sdk.Context, tx sdk.Tx, isSimulate bool) (sdk.Context, error) {
	// all transactions must implement GasTx
	gasTx, ok := tx.(GasTx)
	if !ok {
		// Set a gas meter with limit 0 as to prevent an infinite gas meter attack
		// during runTx.
		newCtx := setGasMeter(ctx, 0, isSimulate)
		return newCtx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be GasTx")
	}

	return setGasMeter(ctx, gasTx.GetGas(), isSimulate), nil
}

// setGasMeter returns a new context with a gas meter set from a given context.
func setGasMeter(ctx sdk.Context, gasLimit uint64, simulate bool) sdk.Context {
	// In various cases such as simulation and during the genesis block, we do not
	// meter any gas utilization.
	if simulate || ctx.BlockHeight() == 0 {
		return ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	}

	return ctx.WithGasMeter(sdk.NewGasMeter(gasLimit))
}
