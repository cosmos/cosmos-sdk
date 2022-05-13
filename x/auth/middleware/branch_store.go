package middleware

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	tmtypes "github.com/tendermint/tendermint/types"
)

// We'll store the original sdk.Context multistore, as well as a
// cache-wrapped multistore, inside the sdk.Context's Values (it's just
// a simple key-value map). We define those keys here, and pass them to the
// middlewares.
var (
	originalMSKey = sdk.ContextKey("ante-original-ms")
	cacheMSKey    = sdk.ContextKey("ante-cache-ms")
)

// WithAnteBranch creates a store branch (cache-wrapped multistore) for Antehandlers.
//
// Usage:
// beginBranch, endBranch := WithAnteBranch()
//
// ComposeMiddlewares(
//   beginBranch,
//   // some middlewares representing antehandlers
//   endBranch, // will write to state before all middlewares below
//   // some other middlewares
// )
func WithAnteBranch() (tx.Middleware, tx.Middleware) {

	beginAnteBranch := func(h tx.Handler) tx.Handler {
		return anteBranchBegin{next: h, originalMSKey: originalMSKey, cacheMSKey: cacheMSKey}
	}
	endAnteBranch := func(h tx.Handler) tx.Handler {
		return anteBranchWrite{next: h, originalMSKey: originalMSKey, cacheMSKey: cacheMSKey}
	}

	return beginAnteBranch, endAnteBranch
}

// anteBranchBegin is the tx.Handler that creates a new branched store.
type anteBranchBegin struct {
	next          tx.Handler
	originalMSKey sdk.ContextKey
	cacheMSKey    sdk.ContextKey
}

// CheckTx implements tx.Handler.CheckTx method.
// Do nothing during CheckTx.
func (sh anteBranchBegin) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	newCtx := anteCreateBranch(ctx, sh.originalMSKey, sh.cacheMSKey, req)

	return sh.next.CheckTx(newCtx, req, checkReq)
}

// DeliverTx implements tx.Handler.DeliverTx method.
func (sh anteBranchBegin) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	newCtx := anteCreateBranch(ctx, sh.originalMSKey, sh.cacheMSKey, req)

	return sh.next.DeliverTx(newCtx, req)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (sh anteBranchBegin) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	newCtx := anteCreateBranch(ctx, sh.originalMSKey, sh.cacheMSKey, req)

	return sh.next.SimulateTx(newCtx, req)
}

// anteCreateBranch creates a new Context based on the existing Context with a MultiStore branch
// in case message processing fails.
func anteCreateBranch(ctx context.Context, originalMSKey, cacheMSKey sdk.ContextKey, req tx.Request) context.Context {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	newSdkCtx, branchedStore := branchStore(sdkCtx, tmtypes.Tx(req.TxBytes))
	// Put the stores inside the new sdk.Context, as we will need them when
	// writing.
	newSdkCtx = newSdkCtx.
		WithValue(cacheMSKey, branchedStore).
		WithValue(originalMSKey, sdkCtx.MultiStore())

	return sdk.WrapSDKContext(newSdkCtx)
}

// branchStore returns a new context based off of the provided context with
// a branched multi-store.
func branchStore(sdkCtx sdk.Context, tx tmtypes.Tx) (sdk.Context, sdk.CacheMultiStore) {
	ms := sdkCtx.MultiStore()
	msCache := ms.CacheMultiStore()
	if msCache.TracingEnabled() {
		msCache = msCache.SetTracingContext(
			sdk.TraceContext(
				map[string]interface{}{
					"txHash": tx.Hash(),
				},
			),
		).(sdk.CacheMultiStore)
	}

	return sdkCtx.WithMultiStore(msCache), msCache
}

// anteBranchWrite is the tx.Handler that commits the state writes of a previously
// created anteBranchBegin tx.Handler.
type anteBranchWrite struct {
	next          tx.Handler
	originalMSKey sdk.ContextKey
	cacheMSKey    sdk.ContextKey
}

// CheckTx implements tx.Handler.CheckTx method.
// Does nothing during CheckTx.
func (sh anteBranchWrite) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	newCtx := anteWrite(ctx, sh.originalMSKey, sh.cacheMSKey)

	return sh.next.CheckTx(newCtx, req, checkReq)
}

// DeliverTx implements tx.Handler.DeliverTx method.
func (sh anteBranchWrite) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	newCtx := anteWrite(ctx, sh.originalMSKey, sh.cacheMSKey)

	return sh.next.DeliverTx(newCtx, req)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (sh anteBranchWrite) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	newCtx := anteWrite(ctx, sh.originalMSKey, sh.cacheMSKey)

	return sh.next.SimulateTx(newCtx, req)
}

// branchAndRun creates a new Context based on the existing Context with a MultiStore branch
// in case message processing fails.
func anteWrite(ctx context.Context, originalMSKey, cacheMSKey sdk.ContextKey) context.Context {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	originalStore := sdkCtx.Value(originalMSKey).(sdk.MultiStore)
	branchedStore := sdkCtx.Value(cacheMSKey).(sdk.CacheMultiStore)

	if !sdkCtx.IsZero() {
		// At this point, newCtx.MultiStore() is a store branch, or something else
		// replaced by the AnteHandler. We want the original multistore.
		//
		// Also, in the case of the tx aborting, we need to track gas consumed via
		// the instantiated gas meter in the AnteHandler, so we update the context
		// prior to returning.
		sdkCtx = sdkCtx.WithMultiStore(originalStore)
	}

	branchedStore.Write()

	// We don't need references to the 2 multistores anymore.
	sdkCtx = sdkCtx.WithValue(originalMSKey, nil).WithValue(cacheMSKey, nil)

	return sdk.WrapSDKContext(sdkCtx)
}

// WithRunMsgsBranch creates a store branch (cache store) for running Msgs.
//
// Usage:
//
// ComposeMiddlewares(
//   // some middlewares
//
//   // Creates a new MultiStore branch, discards downstream writes if the downstream returns error.
//   WithRunMsgsBranch,
//   // optionally, some other middlewares who should also discard writes when
//   // middleware fails.
// )
func WithRunMsgsBranch(h tx.Handler) tx.Handler {
	return runMsgsBranch{next: h}
}

// runMsgsBranch is the tx.Handler that commits the state writes of a previously
// created anteBranchBegin tx.Handler.
type runMsgsBranch struct {
	next tx.Handler
}

// CheckTx implements tx.Handler.CheckTx method.
// Does nothing during CheckTx.
func (sh runMsgsBranch) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	return sh.next.CheckTx(ctx, req, checkReq)
}

// DeliverTx implements tx.Handler.DeliverTx method.
func (sh runMsgsBranch) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	return branchAndRun(ctx, req, sh.next.DeliverTx)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (sh runMsgsBranch) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
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
