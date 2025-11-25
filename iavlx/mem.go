package iavlx

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v4/mem"
)

type MemoryThresholds struct {
	Yellow uint64
	Red    uint64
}

var defaultThresholds = MemoryThresholds{
	Yellow: 2 * 1024 * 1024 * 1024, // 2GB
	Red:    1 * 1024 * 1024 * 1024, // 1GB
}

type memoryController struct {
	evictDepth           atomic.Uint32
	readerUpdateInterval atomic.Uint32
	thresholds           MemoryThresholds
	ctx                  context.Context
	// TODO: memLimit             uint64
}

func NewMemoryController(ctx context.Context) *memoryController {
	mc := &memoryController{
		thresholds: defaultThresholds,
		ctx:        ctx,
	}
	// start with moderate values
	mc.readerUpdateInterval.Store(1)
	mc.evictDepth.Store(8)
	go mc.run()
	return mc
}

func (mc *memoryController) run() {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-mc.ctx.Done():
			return
		case <-ticker.C:
			mc.update()
		}
	}
}

func (mc *memoryController) getMemPressure() (memPressureZone, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		// TODO log
		return memPressureUnknown, fmt.Errorf("failed to get virtual memory info: %w", err)
	}
	available := vm.Available

	switch {
	case available <= mc.thresholds.Red:
		return memPressureRed, nil
	case available <= mc.thresholds.Yellow:
		return memPressureYellow, nil
	default:
		return memPressureGreen, nil
	}
}

func (mc *memoryController) update() {
	zone, err := mc.getMemPressure()
	if err != nil {
		// TODO log
		return
	}

	currentInterval := mc.readerUpdateInterval.Load()
	currentDepth := mc.evictDepth.Load()

	// this allows us to hold up to 1 trillion nodes in memory which is a truly massive amount
	const maxEvictDepth = 40

	// 256 is somewhat arbitrary but seems like a reasonable upper bound for reader update interval
	const maxReaderUpdateInterval = 256

	// This algorithm follows the AIMD (Additive Increase Multiplicative Decrease) strategy
	// with three zones: green, yellow, and red.
	// In the red zone, we aggressively reduce memory usage by setting parameters to minimums.
	// In the yellow zone, we reduce memory usage by dividing parameters.
	// In the green zone, we increase performance by adding to parameters.
	// We also start by first decreasing the reader update interval before the evict depth
	// when we need to reduce memory usage, and vice versa when we can increase performance.
	// The logic for this is:
	// 1. a long reader update interval limits our ability to evict old versions anyway,
	//	and the primary downside is that we need to reopen mmapped files more often
	// 2. a low evict depth limits our ability to keep frequently accessed nodes in memory,
	// 	and this has the biggest impact on performance because all writes are against the latest version
	// 	and most reads are too
	// Thus we prioritize reducing the reader update interval first to free up memory for evictions,
	//	and prioritize increasing the evict depth first to improve performance.
	switch zone {
	case memPressureRed:
		// in the red zone set to minimums ASAP to reduce memory usage quickly and immediately
		mc.readerUpdateInterval.Store(1)
		mc.evictDepth.Store(0)
	case memPressureYellow:
		// in the yellow zone, follow the AIMD strategy and divide to reduce memory usage
		// we start by reducing the reader update interval and then the evict depth
		if currentInterval > 1 {
			mc.readerUpdateInterval.Store(maxUint32(1, currentInterval/2))
		} else if currentDepth > 0 {
			mc.evictDepth.Store(currentDepth / 2)
		}
	case memPressureGreen:
		// in the green zone, follow the AIMD strategy and add to increase performance
		// we start by increasing the evict depth and then the reader update interval
		if currentDepth < maxEvictDepth {
			mc.evictDepth.Store(currentDepth + 1)
		} else if currentInterval < maxReaderUpdateInterval {
			mc.readerUpdateInterval.Store(currentInterval + 1)
		}
	default:
		// unexpected case, do nothing
	}
}

func maxUint32(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func (mc *memoryController) EvictDepth() uint8 {
	return uint8(mc.evictDepth.Load())
}

func (mc *memoryController) ReaderUpdateInterval() uint32 {
	return mc.readerUpdateInterval.Load()
}

type memPressureZone int

const (
	memPressureUnknown memPressureZone = iota
	memPressureGreen
	memPressureYellow
	memPressureRed
)
