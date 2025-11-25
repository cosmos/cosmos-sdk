package iavlx

import (
	"context"
	"runtime"
	"sync/atomic"
	"time"
)

type memoryMonitor struct {
	evictBudget   atomic.Int64
	underPressure func()
	ctx           context.Context
	memoryLimit   uint64
}

func newMemoryMonitor(ctx context.Context, memoryLimit uint64, underPressure func()) *memoryMonitor {
	mc := &memoryMonitor{
		memoryLimit:   memoryLimit,
		ctx:           ctx,
		underPressure: underPressure,
	}
	go mc.run()
	return mc
}

func (mc *memoryMonitor) run() {
	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()

	for {
		select {
		case <-mc.ctx.Done():
			return
		case <-ticker.C:
			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			pressure := int64(mem.Alloc) - int64(mc.memoryLimit)
			if pressure > 0 {
				mc.evictBudget.Store(pressure)
				mc.underPressure()
			}
		}
	}
}

func (mc *memoryMonitor) EvictBudget() *atomic.Int64 {
	return &mc.evictBudget
}

func (mc *memoryMonitor) UnderPressure() bool {
	return mc.evictBudget.Load() > 0
}
