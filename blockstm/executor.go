package blockstm

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
)

// Executor fields are not mutated during execution.
type Executor struct {
	ctx        context.Context // context for cancellation
	scheduler  *Scheduler      // scheduler for task management
	txExecutor TxExecutor      // callback to actually execute a transaction
	mvMemory   *MVMemory       // multi-version memory for the executor

	// index of the executor, used for debugging output
	i int
}

func NewExecutor(
	ctx context.Context,
	scheduler *Scheduler,
	txExecutor TxExecutor,
	mvMemory *MVMemory,
	i int,
) *Executor {
	return &Executor{
		ctx:        ctx,
		scheduler:  scheduler,
		txExecutor: txExecutor,
		mvMemory:   mvMemory,
		i:          i,
	}
}

// Run executes all tasks until completion
// Invariant `num_active_tasks`:
//   - `NextTask` increases it if returns a valid task.
//   - `TryExecute` and `NeedsReexecution` don't change it if it returns a new valid task to run,
//     otherwise it decreases it.
func (e *Executor) Run() error {
	var kind TaskKind
	version := InvalidTxnVersion
	for !e.scheduler.Done() {
		if !version.Valid() {
			// check for cancellation
			select {
			case <-e.ctx.Done():
				return nil
			default:
			}

			version, kind = e.scheduler.NextTask()
			continue
		}

		switch kind {
		case TaskKindExecution:
			version, kind = e.TryExecute(version)
		case TaskKindValidation:
			version, kind = e.NeedsReexecution(version)
		default:
			return fmt.Errorf("unknown task kind %v", kind)
		}
	}
	return nil
}

func (e *Executor) TryExecute(version TxnVersion) (TxnVersion, TaskKind) {
	start := time.Now()
	e.scheduler.executedTxns.Add(1)
	view := e.execute(version.Index)

	// Track read and write counts
	readCount := view.CountReads()
	writeCount := view.CountWrites()
	telemetry.IncrCounter(float32(readCount), TelemetrySubsystem, KeyTxReadCount)   //nolint:staticcheck // TODO: switch to OpenTelemetry
	telemetry.IncrCounter(float32(writeCount), TelemetrySubsystem, KeyTxWriteCount) //nolint:staticcheck // TODO: switch to OpenTelemetry

	wroteNewLocation := e.mvMemory.Record(version, view)
	if wroteNewLocation {
		telemetry.IncrCounter(1, TelemetrySubsystem, KeyTxNewLocationWrite) //nolint:staticcheck // TODO: switch to OpenTelemetry
	}

	telemetry.MeasureSince(start, TelemetrySubsystem, KeyTryExecuteTime)
	return e.scheduler.FinishExecution(version, wroteNewLocation)
}

func (e *Executor) NeedsReexecution(version TxnVersion) (TxnVersion, TaskKind) {
	e.scheduler.validatedTxns.Add(1)
	valid := e.mvMemory.ValidateReadSet(version.Index)
	aborted := !valid && e.scheduler.TryValidationAbort(version)
	if aborted {
		e.mvMemory.ConvertWritesToEstimates(version.Index)
	}
	return e.scheduler.FinishValidation(version.Index, aborted)
}

func (e *Executor) execute(txn TxnIndex) *MultiMVMemoryView {
	view := e.mvMemory.View(txn)
	e.txExecutor(txn, view)
	return view
}
