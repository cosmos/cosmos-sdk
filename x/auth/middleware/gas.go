package middleware

import (
	"context"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// GasTx defines a Tx with a GetGas() method which is needed to use SetUpContextDecorator
type GasTx interface {
	sdk.Tx
	GetGas() uint64
}

type gasTxHandler struct {
	inner tx.TxHandler
}

// NewGasTxMiddleware defines a simple middleware that sets a new GasMeter on
// the sdk.Context. It reads the tx.GetGas() by default, or sets to infinity
// in simulate mode.
func NewGasTxMiddleware() tx.TxMiddleware {
	return func(txh tx.TxHandler) tx.TxHandler {
		return gasTxHandler{inner: txh}
	}
}

var _ tx.TxHandler = gasTxHandler{}

// CheckTx implements TxHandler.CheckTx.
func (txh gasTxHandler) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	sdkCtx, err := gasContext(sdk.UnwrapSDKContext(ctx), tx, false)
	if err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return txh.inner.CheckTx(sdk.WrapSDKContext(sdkCtx), tx, req)
}

// DeliverTx implements TxHandler.DeliverTx.
func (txh gasTxHandler) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	sdkCtx, err := gasContext(sdk.UnwrapSDKContext(ctx), tx, false)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}
	return txh.inner.DeliverTx(sdk.WrapSDKContext(sdkCtx), tx, req)
}

// SimulateTx implements TxHandler.SimulateTx method.
func (txh gasTxHandler) SimulateTx(ctx context.Context, sdkTx sdk.Tx, req tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	sdkCtx, err := gasContext(sdk.UnwrapSDKContext(ctx), sdkTx, true)
	if err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	return txh.inner.SimulateTx(sdk.WrapSDKContext(sdkCtx), sdkTx, req)
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
