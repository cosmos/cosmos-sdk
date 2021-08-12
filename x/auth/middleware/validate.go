package middleware

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
)

type validateTxHandler struct {
	inner tx.TxHandler

	debug bool
}

func validateMiddleware(txHandler tx.TxHandler) tx.TxHandler {
	return validateTxHandler{
		inner: txHandler,
	}
}

var _ tx.TxHandler = validateTxHandler{}

func (txh validateTxHandler) CheckTx(ctx sdk.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	err := validateBasicTxMsgs(tx.GetMsgs())
	if err != nil {
		return sdkerrors.ResponseCheckTx(nil, 0, 0, txh.debug), err
	}

	// Branch context before AnteHandler call in case it aborts.
	// This is required for both CheckTx and DeliverTx.
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/2772
	//
	// NOTE: Alternatively, we could require that AnteHandler ensures that
	// writes do not happen if aborted/failed.  This may have some
	// performance benefits, but it'll be more difficult to get right.
	anteCtx, msCache := txh.cacheTxContext(ctx, req.Tx)
	anteCtx = anteCtx.WithEventManager(sdk.NewEventManager())
	newCtx, err := app.anteHandler(anteCtx, tx, mode == runTxModeSimulate)

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

	return txh.inner.CheckTx(ctx, tx, req)
}

func (txh validateTxHandler) DeliverTx(ctx sdk.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {

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

// cacheTxContext returns a new context based off of the provided context with
// a branched multi-store.
func (txh validateTxHandler) cacheTxContext(ctx sdk.Context, txBytes []byte) (sdk.Context, sdk.CacheMultiStore) {
	ms := ctx.MultiStore()
	// TODO: https://github.com/cosmos/cosmos-sdk/issues/2824
	msCache := ms.CacheMultiStore()
	if msCache.TracingEnabled() {
		msCache = msCache.SetTracingContext(
			sdk.TraceContext(
				map[string]interface{}{
					"txHash": fmt.Sprintf("%X", tmhash.Sum(txBytes)),
				},
			),
		).(sdk.CacheMultiStore)
	}

	return ctx.WithMultiStore(msCache), msCache
}
