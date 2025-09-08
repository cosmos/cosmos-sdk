package blockstm

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"

	storetypes "cosmossdk.io/store/types"
)

func ExecuteBlock(
	ctx context.Context,
	blockSize int,
	stores map[storetypes.StoreKey]int,
	storage MultiStore,
	executors int,
	txExecutor TxExecutor,
) error {
	return ExecuteBlockWithEstimates(
		ctx, blockSize, stores, storage, executors,
		nil, txExecutor,
	)
}

func ExecuteBlockWithEstimates(
	ctx context.Context,
	blockSize int,
	stores map[storetypes.StoreKey]int,
	storage MultiStore,
	executors int,
	estimates []MultiLocations, // txn -> multi-locations
	txExecutor TxExecutor,
) error {
	if executors < 0 {
		return fmt.Errorf("invalid number of executors: %d", executors)
	}
	if executors == 0 {
		executors = maxParallelism()
	}

	// Create a new scheduler
	scheduler := NewScheduler(blockSize)
	mvMemory := NewMVMemoryWithEstimates(blockSize, stores, storage, scheduler, estimates)

	var wg sync.WaitGroup
	wg.Add(executors)
	for i := 0; i < executors; i++ {
		e := NewExecutor(ctx, scheduler, txExecutor, mvMemory, i)
		go func() {
			defer wg.Done()
			e.Run()
		}()
	}
	wg.Wait()

	if !scheduler.Done() {
		if ctx.Err() != nil {
			// cancelled
			return ctx.Err()
		}

		return errors.New("scheduler did not complete")
	}

	// Write the snapshot into the storage
	mvMemory.WriteSnapshot(storage)
	return nil
}

func maxParallelism() int {
	return min(runtime.GOMAXPROCS(0), runtime.NumCPU())
}
