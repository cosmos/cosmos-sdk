package middleware

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	tmtypes "github.com/tendermint/tendermint/types"
)

type branchStoreHandler struct {
	next tx.Handler
}

// WithBranchedStore creates a new MultiStore branch and commits the store if the downstream
// returned no error. It cancels writes from the failed transactions.
func WithBranchedStore(txh tx.Handler) tx.Handler {
	return branchStoreHandler{next: txh}
}

// CheckTx implements tx.Handler.CheckTx method.
// Do nothing during CheckTx.
func (sh branchStoreHandler) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	return sh.next.CheckTx(ctx, req, checkReq)
}

// DeliverTx implements tx.Handler.DeliverTx method.
func (sh branchStoreHandler) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	return branchAndRun(ctx, req, sh.next.DeliverTx)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (sh branchStoreHandler) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	return branchAndRun(ctx, req, sh.next.SimulateTx)
}

type nextFn func(ctx context.Context, req tx.Request) (tx.Response, error)

// branchAndRun creates a new Context based on the existing Context with a MultiStore branch
// in case message processing fails.
func branchAndRun(ctx context.Context, req tx.Request, fn nextFn) (tx.Response, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	runMsgCtx, branchedStore := branchStore(sdkCtx, tmtypes.Tx(req.TxBytes))

	rsp, err := fn(sdk.WrapSDKContext(runMsgCtx), req)
	if err == nil {
		// commit storage iff no error
		branchedStore.Write()
	}

	return rsp, err
}

// branchStore returns a new context based off of the provided context with
// a branched multi-store.
func branchStore(sdkCtx sdk.Context, tx tmtypes.Tx) (sdk.Context, sdk.CacheMultiStore) {
	ms := sdkCtx.MultiStore()
	msCache := ms.CacheWrap()
	if msCache.TracingEnabled() {
		msCache.SetTracingContext(
			sdk.TraceContext(
				map[string]interface{}{
					"txHash": tx.Hash(),
				},
			),
		)
	}

	return sdkCtx.WithMultiStore(msCache), msCache
}
