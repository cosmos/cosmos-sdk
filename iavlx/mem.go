package iavlx

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
)

type memoryMonitor struct {
	evictBudget atomic.Int64
	mu          sync.Mutex
	cond        *sync.Cond
	ctx         context.Context
	logger      *slog.Logger
}

func newMemoryMonitor(ctx context.Context, logger *slog.Logger, memoryLimit uint64) *memoryMonitor {
	mc := &memoryMonitor{
		ctx:    ctx,
		logger: logger,
	}
	mc.evictBudget.Store(-int64(memoryLimit))
	mc.cond = sync.NewCond(&mc.mu)

	logger.InfoContext(ctx, "memoryMonitor initialized", "memoryLimit", memoryLimit, "initialBudget", -int64(memoryLimit))

	// Broadcast on context cancellation to wake waiting evictors
	go func() {
		<-ctx.Done()
		mc.cond.Broadcast()
	}()

	return mc
}

// Wait blocks until there is memory pressure or the context is cancelled.
// Returns true if there is work to do, false if cancelled.
func (mc *memoryMonitor) Wait() bool {
	mc.mu.Lock()
	for mc.evictBudget.Load() <= 0 && mc.ctx.Err() == nil {
		mc.cond.Wait()
	}
	mc.mu.Unlock()
	return mc.ctx.Err() == nil
}

func (mc *memoryMonitor) UnderPressure() bool {
	return mc.evictBudget.Load() > 0
}

func (mc *memoryMonitor) AddUsage(usage uint64) {
	before := mc.evictBudget.Load()
	pressure := mc.evictBudget.Add(int64(usage))
	mc.logger.DebugContext(mc.ctx, "AddUsage", "usage", usage, "budgetBefore", before, "budgetAfter", pressure, "underPressure", pressure > 0)
	if pressure > 0 {
		mc.cond.Broadcast()
	}
}

// TrackEviction tracks memory that has been dereferenced due to eviction and returns true if eviction should continue.
func (mc *memoryMonitor) TrackEviction(memNode *MemNode) bool {
	var sz int
	if memNode.IsLeaf() {
		sz = SizeLeaf + len(memNode.key) + len(memNode.value)
	} else {
		sz = SizeBranch + len(memNode.key)
	}
	return mc.evictBudget.Add(int64(-sz)) >= 0
}
