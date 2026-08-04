package baseapp

import (
	"context"

	"github.com/cockroachdb/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
)

// mempoolRemoveQueueSize is the capacity of the buffered channel used to
// enqueue mempool removals. Mempool removal is best-effort: when the queue is
// full, further removals are dropped (and logged) instead of blocking the
// consensus-critical path.
const mempoolRemoveQueueSize = 1024

// mempoolRemoveRequest is a single mempool removal enqueued for asynchronous
// processing by the mempool removal worker.
type mempoolRemoveRequest struct {
	tx     sdk.Tx
	reason mempool.RemoveReason
}

// startMempoolRemoveWorker spawns the goroutine that drains mempoolRemoveCh and
// performs mempool removals outside of the consensus-critical path.
func (app *BaseApp) startMempoolRemoveWorker() {
	app.mempoolRemoveWg.Add(1)
	go func() {
		defer app.mempoolRemoveWg.Done()
		for req := range app.mempoolRemoveCh {
			app.doMempoolRemove(req.tx, req.reason)
		}
	}()
}

// enqueueMempoolRemove queues a mempool removal to be processed asynchronously.
//
// Mempool state is node-local and may legitimately diverge between nodes (a tx
// may already have been removed, a custom mempool may reject the tx, or a
// mempool implementation may even panic while removing). Such errors or panics
// must never propagate into block execution, where they would cause a consensus
// failure (app hash mismatch). This method therefore never blocks and never
// returns an error: the actual removal is performed by a background worker
// which recovers from panics and logs errors.
//
// Note: because removal is asynchronous, a tx that is re-inserted into the
// mempool (e.g. by a concurrent CheckTx) before its queued removal is processed
// may be removed from the pool as a side effect. This only affects the
// node-local mempool, never the state machine, and stale txs are re-inserted by
// the next recheck.
func (app *BaseApp) enqueueMempoolRemove(tx sdk.Tx, reason mempool.RemoveReason) {
	select {
	case app.mempoolRemoveCh <- mempoolRemoveRequest{tx: tx, reason: reason}:
	default:
		// The queue is full: drop the removal rather than stalling block
		// execution. Stale txs are pruned by the next recheck.
		app.logger.Warn("mempool remove queue is full; dropping removal", "caller", reason.Caller)
	}
}

// doMempoolRemove performs a single mempool removal, recovering from panics and
// logging errors so that a faulty mempool can never fail the node's consensus.
func (app *BaseApp) doMempoolRemove(tx sdk.Tx, reason mempool.RemoveReason) {
	defer func() {
		if r := recover(); r != nil {
			app.logger.Error("panic while removing tx from mempool; ignoring", "panic", r, "caller", reason.Caller)
		}
	}()

	// Use a fresh background context: the sdk.Context captured during block
	// execution must not be reused outside of it.
	if err := mempool.RemoveWithReason(context.Background(), app.mempool, tx, reason); err != nil && !errors.Is(err, mempool.ErrTxNotFound) {
		app.logger.Error("failed to remove tx from mempool; ignoring", "err", err, "caller", reason.Caller)
	}
}
