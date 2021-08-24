package middleware

import (
	"context"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

type legacyAnteTxHandler struct {
	anteHandler sdk.AnteHandler
	inner       tx.Handler
}

func newLegacyAnteMiddleware(anteHandler sdk.AnteHandler) tx.Middleware {
	return func(txHandler tx.Handler) tx.Handler {
		return legacyAnteTxHandler{
			anteHandler: anteHandler,
			inner:       txHandler,
		}
	}
}

var _ tx.Handler = legacyAnteTxHandler{}

// CheckTx implements tx.Handler.CheckTx method.
func (txh legacyAnteTxHandler) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	sdkCtx, err := txh.runAnte(ctx, tx, req.Tx, false)
	if err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return txh.inner.CheckTx(sdk.WrapSDKContext(sdkCtx), tx, req)
}

// DeliverTx implements tx.Handler.DeliverTx method.
func (txh legacyAnteTxHandler) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	sdkCtx, err := txh.runAnte(ctx, tx, req.Tx, false)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return txh.inner.DeliverTx(sdk.WrapSDKContext(sdkCtx), tx, req)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (txh legacyAnteTxHandler) SimulateTx(ctx context.Context, sdkTx sdk.Tx, req tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	sdkCtx, err := txh.runAnte(ctx, sdkTx, req.TxBytes, true)
	if err != nil {
		return tx.ResponseSimulateTx{}, err
	}

	return txh.inner.SimulateTx(sdk.WrapSDKContext(sdkCtx), sdkTx, req)
}

func (txh legacyAnteTxHandler) runAnte(ctx context.Context, tx sdk.Tx, txBytes []byte, isSimulate bool) (sdk.Context, error) {
	err := validateBasicTxMsgs(tx.GetMsgs())
	if err != nil {
		return sdk.Context{}, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if txh.anteHandler == nil {
		return sdkCtx, nil
	}

	ms := sdkCtx.MultiStore()

	// Branch context before AnteHandler call in case it aborts.
	// This is required for both CheckTx and DeliverTx.
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/2772
	//
	// NOTE: Alternatively, we could require that AnteHandler ensures that
	// writes do not happen if aborted/failed.  This may have some
	// performance benefits, but it'll be more difficult to get right.
	anteCtx, msCache := cacheTxContext(sdkCtx, txBytes)
	anteCtx = anteCtx.WithEventManager(sdk.NewEventManager())
	newCtx, err := txh.anteHandler(anteCtx, tx, isSimulate)
	if err != nil {
		return sdk.Context{}, err
	}

	if !newCtx.IsZero() {
		// At this point, newCtx.MultiStore() is a store branch, or something else
		// replaced by the AnteHandler. We want the original multistore.
		//
		// Also, in the case of the tx aborting, we need to track gas consumed via
		// the instantiated gas meter in the AnteHandler, so we update the context
		// prior to returning.
		sdkCtx = newCtx.WithMultiStore(ms)
	}

	msCache.Write()

	return sdkCtx, nil
}

// validateBasicTxMsgs executes basic validator calls for messages.
func validateBasicTxMsgs(msgs []sdk.Msg) error {
	if len(msgs) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "must contain at least one message")
	}

	for _, msg := range msgs {
		err := msg.ValidateBasic()
		if err != nil {
			return err
		}
	}

	return nil
}
