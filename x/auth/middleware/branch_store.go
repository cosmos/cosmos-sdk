package middleware

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	tmtypes "github.com/tendermint/tendermint/types"
)

// WithBranchAnte creates a store branch (cache store) for Antehandlers.
//
// Usage:
// beginBranch, endBranch := WithBranchAnte()
//
// ComposeMiddlewares(
//   beginBranch,
//   // some middlewares representing antehandlers
//   endBranch, // will write to state before all middlewares below
//   // some other middlewares
// )
func WithBranchAnte() (tx.Middleware, tx.Middleware) {
	return withBranchStore("ante", false)
}

// WithBranchAnte creates a store branch (cache store) for running Msgs.
//
// Usage:
// beginBranch, endBranch := WithBranchRunMsgs()
//
// ComposeMiddlewares(
//   // some middlewares
//   beginBranch,
//   // some middlewares that will not write to state if they fail
//   endBranch, // will write to state right after runMsgs
// )
func WithBranchRunMsgs() (tx.Middleware, tx.Middleware) {
	return withBranchStore("runMsgs", true)
}

// withBranchStore creates 2 middlewares:
// - one that creates a branched store (i.e. a cache store),
// - another that writes the cache store to its parent store.
//
// For the writer middleware, we pass a `writeAfterNext` to decide if we write
// before or after the `next` tx.Handler.
func withBranchStore(branchName string, writeAfterNext bool) (tx.Middleware, tx.Middleware) {
	key := sdk.ContextKey(branchName)

	return func(h tx.Handler) tx.Handler {
			return branchBegin{next: h, branchName: key}
		}, func(h tx.Handler) tx.Handler {
			return branchWrite{next: h, branchName: key, writeAfterNext: writeAfterNext}
		}
}

// branchBegin is the tx.Handler that creates a new branched store.
type branchBegin struct {
	next       tx.Handler
	branchName sdk.ContextKey
}

// CheckTx implements tx.Handler.CheckTx method.
// Do nothing during CheckTx.
func (sh branchBegin) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	return sh.next.CheckTx(ctx, req, checkReq)
}

// DeliverTx implements tx.Handler.DeliverTx method.
func (sh branchBegin) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	return createBranch(ctx, sh.branchName, req, sh.next.DeliverTx)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (sh branchBegin) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	return createBranch(ctx, sh.branchName, req, sh.next.SimulateTx)
}

type nextFn func(ctx context.Context, req tx.Request) (tx.Response, error)

// createBranch creates a new Context based on the existing Context with a MultiStore branch
// in case message processing fails.
func createBranch(ctx context.Context, branchName sdk.ContextKey, req tx.Request, fn nextFn) (tx.Response, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	newSdkCtx, branchedStore := branchStore(sdkCtx, tmtypes.Tx(req.TxBytes))

	newSdkCtx = newSdkCtx.WithValue(branchName, branchedStore)
	newCtx := sdk.WrapSDKContext(newSdkCtx)

	return fn(newCtx, req)
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

// branchWrite is the tx.Handler that commits the state writes of a previously
// created branchBegin tx.Handler.
type branchWrite struct {
	next           tx.Handler
	branchName     sdk.ContextKey
	writeAfterNext bool
}

// CheckTx implements tx.Handler.CheckTx method.
// Do nothing during CheckTx.
func (sh branchWrite) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	return sh.next.CheckTx(ctx, req, checkReq)
}

// DeliverTx implements tx.Handler.DeliverTx method.
func (sh branchWrite) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	branchedStore := ctx.Value(sh.branchName).(sdk.CacheMultiStore)
	if !sh.writeAfterNext {
		branchedStore.Write()
	}

	res, err := sh.next.DeliverTx(ctx, req)

	if sh.writeAfterNext {
		branchedStore.Write()
	}

	return res, err
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (sh branchWrite) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	branchedStore := ctx.Value(sh.branchName).(sdk.CacheMultiStore)
	if !sh.writeAfterNext {
		branchedStore.Write()
	}

	res, err := sh.next.SimulateTx(ctx, req)

	if sh.writeAfterNext {
		branchedStore.Write()
	}

	return res, err
}
