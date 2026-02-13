package blockstm

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"golang.org/x/sync/errgroup"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
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
	_, _, err := executeBlockWithEstimatesImpl(
		ctx,
		blockSize,
		stores,
		storage,
		executors,
		estimates,
		txExecutor,
		true,
	)
	return err
}

func executeBlockWithEstimatesImpl(
	ctx context.Context,
	blockSize int,
	stores map[storetypes.StoreKey]int,
	storage MultiStore,
	executors int,
	estimates []MultiLocations, // txn -> multi-locations
	txExecutor TxExecutor,
	emitTelemetry bool,
) (executed, validated uint64, err error) {
	if executors < 0 {
		return 0, 0, fmt.Errorf("invalid number of executors: %d", executors)
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

	err = wg.Wait()
	close(cancelDone)
	if err != nil {
		return 0, 0, err
	}

	if !scheduler.Done() {
		if ctx.Err() != nil {
			return 0, 0, ctx.Err()
		}
		return 0, 0, errors.New("scheduler did not complete")
	}

	executed = uint64(scheduler.executedTxns.Load())
	validated = uint64(scheduler.validatedTxns.Load())

	if emitTelemetry {
		telemetry.SetGauge(float32(executed), TelemetrySubsystem, KeyExecutedTxs)   //nolint:staticcheck // TODO: switch to OpenTelemetry
		telemetry.SetGauge(float32(validated), TelemetrySubsystem, KeyValidatedTxs) //nolint:staticcheck // TODO: switch to OpenTelemetry
	}

	// Write the snapshot into the storage
	mvMemory.WriteSnapshot(storage)
	return executed, validated, nil
}

func maxParallelism() int {
	return min(runtime.GOMAXPROCS(0), runtime.NumCPU())
}
