package iavlx

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v4/mem"
)

type memoryMonitor struct {
	hasPressure  atomic.Bool
	memThreshold uint64
	mu           sync.Mutex
	cond         *sync.Cond
	evictBudget  *atomic.Int64
	ctx          context.Context
	// TODO: memLimit             uint64
}

func newMemoryMonitor(ctx context.Context) *memoryMonitor {
	const defaultMemThreshold = 2 * 1024 * 1024 * 1024 // 2 GB
	mc := &memoryMonitor{
		memThreshold: defaultMemThreshold,
		ctx:          ctx,
	}
	mc.cond = sync.NewCond(&mc.mu)
	go mc.run()
	return mc
}

func (mc *memoryMonitor) HasPressure() bool {
	return mc.hasPressure.Load()
}

func (mc *memoryMonitor) Wait() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.cond.Wait()
}

func (mc *memoryMonitor) run() {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	wasPressure := false
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
			isPressure := vm.Available <= mc.memThreshold
			mc.hasPressure.Store(isPressure)

			if isPressure && !wasPressure {
				mc.cond.Broadcast()
			}
			wasPressure = isPressure
		}
	}
}

func (mc *memoryMonitor) EvictBudget() *atomic.Int64 {
	return mc.evictBudget
}
