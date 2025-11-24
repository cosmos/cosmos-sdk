package blockstm

import (
	"context"
	"fmt"
)

// TxExecutor defines the callback type for executing a single transaction.
// It takes the transaction index and a Multi-Version Memory View.
type TxExecutor func(txn TxnIndex, view *MultiMVMemoryView)

// Executor is responsible for running execution and validation tasks received from the Scheduler.
// Its fields are designed to be read-only (not mutated) during the concurrent execution phase.
type Executor struct {
	ctx             context.Context // Context for external cancellation signals
	scheduler       *Scheduler      // Central scheduler for task and state management
	txExecutor      TxExecutor      // The function that performs the actual transaction logic
	mvMemory        *MVMemory       // Multi-Version Memory layer for optimistic execution

	// Unique index of the executor, primarily used for logging/debugging
	index int 
}

// NewExecutor creates a new Executor instance.
func NewExecutor(
	ctx context.Context,
	scheduler *Scheduler,
	txExecutor TxExecutor,
	mvMemory *MVMemory,
	executorIndex int, // Renamed 'i' to 'executorIndex' for clarity
) *Executor {
	return &Executor{
		ctx:             ctx,
		scheduler:       scheduler,
		txExecutor:      txExecutor,
		mvMemory:        mvMemory,
		index:           executorIndex,
	}
}

// Run executes all tasks assigned by the Scheduler until the block processing is complete.
// The loop continuously fetches tasks until Scheduler.Done() returns true.
func (e *Executor) Run() error {
	var kind TaskKind
	version := InvalidTxnVersion
	
	// Loop until the scheduler signals completion (i.e., all transactions processed)
	for !e.scheduler.Done() {
		// If no valid task version is currently held, fetch the next one.
		if !version.Valid() {
			// Non-blocking check for context cancellation before fetching a new task
			select {
			case <-e.ctx.Done():
				// Return nil if execution was stopped by cancellation
				return nil 
			default:
			}

			// Get the next task from the scheduler
			version, kind = e.scheduler.NextTask()
			continue
		}

		// Process the fetched task based on its kind
		switch kind {
		case TaskKindExecution:
			// Attempt execution phase
			version, kind = e.TryExecute(version)
		case TaskKindValidation:
			// Attempt validation phase
			version, kind = e.NeedsReexecution(version)
		default:
			return fmt.Errorf("unknown task kind %v", kind)
		}
	}
	return nil
}

// TryExecute performs the execution phase for a given transaction version.
func (e *Executor) TryExecute(version TxnVersion) (TxnVersion, TaskKind) {
	// Increment the executed transaction counter (assumed to be thread-safe)
	e.scheduler.executedTxns.Add(1) 
	
	// Execute the transaction logic using the multi-version view
	view := e.execute(version.Index)
	
	// Record the write set and check for write conflicts
	wroteNewLocation := e.mvMemory.Record(version, view)
	
	// Signal the scheduler that execution is complete, and get the next task
	return e.scheduler.FinishExecution(version, wroteNewLocation)
}

// NeedsReexecution performs the validation phase for a given transaction version.
func (e *Executor) NeedsReexecution(version TxnVersion) (TxnVersion, TaskKind) {
	// Increment the validated transaction counter
	e.scheduler.validatedTxns.Add(1)
	
	// Check if the read set is still valid against current memory state
	valid := e.mvMemory.ValidateReadSet(version.Index)
	
	// Try to abort the transaction if validation failed.
	// Only the scheduler can definitively decide to abort/re-execute.
	aborted := !valid && e.scheduler.TryValidationAbort(version)
	
	if aborted {
		// If aborted, prepare the memory for re-execution by converting writes to estimates
		e.mvMemory.ConvertWritesToEstimates(version.Index)
	}
	
	// Signal the scheduler that validation is complete, and get the next task
	return e.scheduler.FinishValidation(version.Index, aborted)
}

// execute is a helper function to run the TxExecutor callback and return the view.
func (e *Executor) execute(txn TxnIndex) *MultiMVMemoryView {
	view := e.mvMemory.View(txn)
	// The transaction logic populates the view's read/write sets
	e.txExecutor(txn, view) 
	return view
}
