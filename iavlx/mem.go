package iavlx

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v4/mem"
)

type memoryMonitor struct {
	memThreshold  uint64
	evictBudget   atomic.Int64
	underPressure func()
	ctx           context.Context
	// TODO: memLimit             uint64
}

func newMemoryMonitor(ctx context.Context, underPressure func()) *memoryMonitor {
	const defaultMemThreshold = 2 * 1024 * 1024 * 1024 // 2 GB
	mc := &memoryMonitor{
		memThreshold:  defaultMemThreshold,
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
			vm, err := mem.VirtualMemory()
			if err != nil {
				// TODO log
				continue
			}
			pressure := int64(mc.memThreshold) - int64(vm.Available)
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
