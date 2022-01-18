package baseapp_test

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

type handlerFun func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error)

type customTxHandler struct {
	handler handlerFun
	next    tx.Handler
}

var _ tx.Handler = customTxHandler{}

// CustomTxMiddleware is being used in tests for testing
// custom pre-`runMsgs` logic (also called antehandlers before).
func CustomTxHandlerMiddleware(handler handlerFun) tx.Middleware {
	return func(txHandler tx.Handler) tx.Handler {
		return customTxHandler{
			handler: handler,
			next:    txHandler,
		}
	}
}

// CheckTx implements tx.Handler.CheckTx method.
func (txh customTxHandler) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	sdkCtx, err := txh.runHandler(ctx, req.Tx, req.TxBytes, false)
	if err != nil {
		return tx.Response{}, tx.ResponseCheckTx{}, err
	}

	return txh.next.CheckTx(sdk.WrapSDKContext(sdkCtx), req, checkReq)
}

// DeliverTx implements tx.Handler.DeliverTx method.
func (txh customTxHandler) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	sdkCtx, err := txh.runHandler(ctx, req.Tx, req.TxBytes, false)
	if err != nil {
		return tx.Response{}, err
	}

	return txh.next.DeliverTx(sdk.WrapSDKContext(sdkCtx), req)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (txh customTxHandler) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	sdkCtx, err := txh.runHandler(ctx, req.Tx, req.TxBytes, true)
	if err != nil {
		return tx.Response{}, err
	}

	return txh.next.SimulateTx(sdk.WrapSDKContext(sdkCtx), req)
}

func (txh customTxHandler) runHandler(ctx context.Context, tx sdk.Tx, txBytes []byte, isSimulate bool) (sdk.Context, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if txh.handler == nil {
		return sdkCtx, nil
	}

	store := sdkCtx.MultiStore()

	// Branch context before Handler call in case it aborts.
	// This is required for both CheckTx and DeliverTx.
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/2772
	//
	// NOTE: Alternatively, we could require that Handler ensures that
	// writes do not happen if aborted/failed.  This may have some
	// performance benefits, but it'll be more difficult to get right.
	cacheCtx, storeCache := cacheTxContext(sdkCtx, txBytes)
	cacheCtx = cacheCtx.WithEventManager(sdk.NewEventManager())
	newCtx, err := txh.handler(cacheCtx, tx, isSimulate)
	if err != nil {
		return sdk.Context{}, err
	}

	if !newCtx.IsZero() {
		// At this point, newCtx.MultiStore() is a store branch, or something else
		// replaced by the Handler. We want the original multistore.
		//
		// Also, in the case of the tx aborting, we need to track gas consumed via
		// the instantiated gas meter in the Handler, so we update the context
		// prior to returning.
		sdkCtx = newCtx.WithMultiStore(store)
	}

	storeCache.Write()

	return sdkCtx, nil
}

// cacheTxContext returns a new context based off of the provided context with
// a branched multi-store.
func cacheTxContext(sdkCtx sdk.Context, txBytes []byte) (sdk.Context, sdk.CacheMultiStore) {
	store := sdkCtx.MultiStore()
	// TODO: https://github.com/cosmos/cosmos-sdk/issues/2824
	storeCache := store.CacheWrap()
	if storeCache.TracingEnabled() {
		storeCache.SetTracingContext(
			sdk.TraceContext(
				map[string]interface{}{
					"txHash": fmt.Sprintf("%X", tmhash.Sum(txBytes)),
				},
			),
		)
	}

	return sdkCtx.WithMultiStore(storeCache), storeCache
}
