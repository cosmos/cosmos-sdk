package blockstm

import (
	"context"
	"fmt"
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
	var (
		kind TaskKind
		wave Wave // invariant: wave must be set if kind is TaskKindValidation
	)
	version := InvalidTxnVersion
	for !e.scheduler.Done() {
		if !version.Valid() {
			// check for cancellation
			select {
			case <-e.ctx.Done():
				return nil
			default:
			}

			e.ProcessCommits()

			version, wave, kind = e.scheduler.NextTask()
			continue
		}

		switch kind {
		case TaskKindExecution:
			version, wave, kind = e.TryExecute(version)
		case TaskKindValidation:
			version, kind = e.NeedsReexecution(version, wave)
		default:
			return fmt.Errorf("unknown task kind %v", kind)
		}
	}
	return nil
}

func (e *Executor) TryExecute(version TxnVersion) (TxnVersion, Wave, TaskKind) {
	e.scheduler.executedTxns.Add(1)
	view := e.execute(version.Index)
	wroteNewLocation := e.mvMemory.Record(version, view)
	return e.scheduler.FinishExecution(version, wroteNewLocation)
}

func (e *Executor) NeedsReexecution(version TxnVersion, wave Wave) (TxnVersion, TaskKind) {
	e.scheduler.validatedTxns.Add(1)
	valid := e.mvMemory.ValidateReadSet(version.Index)

	var aborted bool
	if !valid {
		// validations on the same transaction can be run concurrently,
		// but only one executor can abort it.
		aborted = e.scheduler.TryValidationAbort(version)
	}

	if aborted {
		e.mvMemory.ConvertWritesToEstimates(version.Index)
	}
	return e.scheduler.FinishValidation(version.Index, wave, aborted, valid)
}

func (e *Executor) execute(txn TxnIndex) *MultiMVMemoryView {
	view := e.mvMemory.View(txn)
	e.txExecutor(txn, view)
	return view
}

func (e *Executor) ProcessCommits() {
	// keep processing if there's work to do and we can acquire the lock
	for e.scheduler.CommitTryLock() {
		for {
			txn, incarnation, ok := e.scheduler.TryCommit()
			if !ok {
				break
			}

			// validate delayed read descriptors
			valid := e.mvMemory.ValidateDelayedReadSet(txn)
			if !valid {
				// re-execute the tx immediately
				version := TxnVersion{txn, incarnation + 1}
				view := e.execute(txn)
				e.mvMemory.Record(version, view)

				e.scheduler.DecreaseValidationIdx(txn + 1)
			}

			// TODO materialize deltas
		}

		e.scheduler.CommitUnlock()
	}
}
