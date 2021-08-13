package middleware

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
)

type legacyAnteTxHandler struct {
	anteHandler sdk.AnteHandler
	inner       tx.TxHandler
}

func newLegacyAnteMiddleware(anteHandler sdk.AnteHandler) tx.TxMiddleware {
	return func(txHandler sdktx.TxHandler) sdktx.TxHandler {
		return legacyAnteTxHandler{
			anteHandler: anteHandler,
			inner:       txHandler,
		}
	}
}

var _ tx.TxHandler = legacyAnteTxHandler{}

func (txh legacyAnteTxHandler) CheckTx(ctx sdk.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	if err := txh.runAnte(ctx, tx, req.Tx); err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return txh.inner.CheckTx(ctx, tx, req)
}

func (txh legacyAnteTxHandler) DeliverTx(ctx sdk.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	if err := txh.runAnte(ctx, tx, req.Tx); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return txh.inner.DeliverTx(ctx, tx, req)
}

func (txh legacyAnteTxHandler) runAnte(ctx sdk.Context, tx sdk.Tx, txBytes []byte) error {
	err := validateBasicTxMsgs(tx.GetMsgs())
	if err != nil {
		return err
	}

	ms := ctx.MultiStore()

	// Branch context before AnteHandler call in case it aborts.
	// This is required for both CheckTx and DeliverTx.
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/2772
	//
	// NOTE: Alternatively, we could require that AnteHandler ensures that
	// writes do not happen if aborted/failed.  This may have some
	// performance benefits, but it'll be more difficult to get right.
	anteCtx, msCache := cacheTxContext(ctx, txBytes)
	newCtx, err := txh.anteHandler(anteCtx, tx, false)
	if err != nil {
		return err
	}

	if !newCtx.IsZero() {
		// At this point, newCtx.MultiStore() is a store branch, or something else
		// replaced by the AnteHandler. We want the original multistore.
		//
		// Also, in the case of the tx aborting, we need to track gas consumed via
		// the instantiated gas meter in the AnteHandler, so we update the context
		// prior to returning.
		ctx = newCtx.WithMultiStore(ms)
	}

	msCache.Write()

	return nil
}

// validateBasicTxMsgs executes basic validator calls for messages.
func validateBasicTxMsgs(msgs []sdk.Msg) error {
	if len(msgs) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "must contain at least one message")
	}

	for _, msg := range msgs {
		err := sdktx.ValidateMsg(msg)
		if err != nil {
			return err
		}
	}

	return nil
}
