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
	if executors < 0 {
		return fmt.Errorf("invalid number of executors: %d", executors)
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
	if err := wg.Wait(); err != nil {
		return err
	}

	if !scheduler.Done() {
		if ctx.Err() != nil {
			// canceled
			return ctx.Err()
		}

		return errors.New("scheduler did not complete")
	}

	telemetry.SetGauge(float32(scheduler.executedTxns.Load()), TelemetrySubsystem, KeyExecutedTxs)   //nolint:staticcheck // TODO: switch to OpenTelemetry
	telemetry.SetGauge(float32(scheduler.validatedTxns.Load()), TelemetrySubsystem, KeyValidatedTxs) //nolint:staticcheck // TODO: switch to OpenTelemetry

	// Write the snapshot into the storage
	mvMemory.WriteSnapshot(storage)
	return nil
}

func maxParallelism() int {
	return min(runtime.GOMAXPROCS(0), runtime.NumCPU())
}
