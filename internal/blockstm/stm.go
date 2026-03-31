package blockstm

import (
	"context"
	"errors"
	"fmt"
	"math"
	"runtime"

	"golang.org/x/sync/errgroup"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
)

func ExecuteBlock(
	ctx context.Context,
	blockSize int,
	stores map[storetypes.StoreKey]int,
	storage MultiStore,
	executors int,
	txExecutor TxExecutor,
) (*BlockExecutionDebug, error) {
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
) (*BlockExecutionDebug, error) {
	if blockSize > math.MaxUint32 {
		return nil, fmt.Errorf("block size overflows uint32: %d", blockSize)
	}

	if executors < 0 {
		return nil, fmt.Errorf("invalid number of executors: %d", executors)
	}
	if executors == 0 {
		executors = maxParallelism()
	}

	// Create a new scheduler
	scheduler := NewScheduler(blockSize)
	mvMemory := NewMVMemoryWithEstimates(blockSize, stores, storage, scheduler, estimates)

	// var wg sync.WaitGroup
	var wg errgroup.Group
	wg.SetLimit(executors)
	for i := 0; i < executors; i++ {
		e := NewExecutor(ctx, scheduler, txExecutor, mvMemory, i)
		wg.Go(e.Run)
	}

	// wake up suspended executors when context is canceled to prevent hanging
	cancelDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			scheduler.CancelAll(func(i TxnIndex) {
				// clear estimates before waking up so they don't suspend again
				mvMemory.ClearEstimates(i)
			})
		case <-cancelDone:
		}
	}()

	err := wg.Wait()
	close(cancelDone)
	debug := scheduler.Debug()
	if err != nil {
		return debug, err
	}

	if !scheduler.Done() {
		if ctx.Err() != nil {
			return debug, ctx.Err()
		}
		return debug, errors.New("scheduler did not complete")
	}

	if inst != nil {
		inst.ExecutedTxs.Add(ctx, scheduler.executedTxns.Load())
		inst.ValidatedTxs.Add(ctx, scheduler.validatedTxns.Load())
		inst.DecreaseCount.Add(ctx, int64(scheduler.decreaseCnt.Load()))
		if blockSize > 0 {
			inst.ExecutionRatio.Add(ctx, float64(scheduler.executedTxns.Load())/float64(blockSize))
		}
	}

	// Write the snapshot into the storage
	mvMemory.WriteSnapshot(ctx, storage)
	return debug, nil
}

func maxParallelism() int {
	return min(runtime.GOMAXPROCS(0), runtime.NumCPU())
}
