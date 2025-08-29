package blockstm

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

type TaskKind int

const (
	TaskKindExecution TaskKind = iota
	TaskKindValidation
)

type TxDependency struct {
	sync.Mutex
	dependents []TxnIndex
}

func (t *TxDependency) Swap(new []TxnIndex) []TxnIndex {
	t.Lock()
	old := t.dependents
	t.dependents = new
	t.Unlock()
	return old
}

// Scheduler implements the scheduler for the block-stm
// ref: `Algorithm 4 The Scheduler module, variables, utility APIs and next task logic`
type Scheduler struct {
	block_size int

	// An index that tracks the next transaction to try and execute.
	execution_idx atomic.Uint64
	// A similar index for tracking validation.
	validation_idx atomic.Uint64
	// Number of times validation_idx or execution_idx was decreased
	decrease_cnt atomic.Uint64
	// Number of ongoing validation and execution tasks
	num_active_tasks atomic.Uint64
	// Marker for completion
	done_marker atomic.Bool

	// txn_idx to a mutex-protected set of dependent transaction indices
	txn_dependency []TxDependency
	// txn_idx to a mutex-protected pair (incarnation_number, status), where status ∈ {READY_TO_EXECUTE, EXECUTING, EXECUTED, ABORTING}.
	txn_status []StatusEntry

	// metrics
	executedTxns  atomic.Int64
	validatedTxns atomic.Int64
}

func NewScheduler(block_size int) *Scheduler {
	return &Scheduler{
		block_size:     block_size,
		txn_dependency: make([]TxDependency, block_size),
		txn_status:     make([]StatusEntry, block_size),
	}
}

func (s *Scheduler) Done() bool {
	return s.done_marker.Load()
}

func (s *Scheduler) DecreaseValidationIdx(target TxnIndex) {
	StoreMin(&s.validation_idx, uint64(target))
	s.decrease_cnt.Add(1)
}

func (s *Scheduler) CheckDone() {
	observed_cnt := s.decrease_cnt.Load()
	if s.execution_idx.Load() >= uint64(s.block_size) &&
		s.validation_idx.Load() >= uint64(s.block_size) &&
		s.num_active_tasks.Load() == 0 {
		if observed_cnt == s.decrease_cnt.Load() {
			s.done_marker.Store(true)
		}
	}
	// avoid busy waiting
	runtime.Gosched()
}

// TryIncarnate tries to incarnate a transaction index to execute.
// Returns the transaction version if successful, otherwise returns invalid version.
//
// Invariant `num_active_tasks`: decreased if an invalid task is returned.
func (s *Scheduler) TryIncarnate(idx TxnIndex) TxnVersion {
	if int(idx) < s.block_size {
		if incarnation, ok := s.txn_status[idx].TrySetExecuting(); ok {
			return TxnVersion{idx, incarnation}
		}
	}
	DecrAtomic(&s.num_active_tasks)
	return InvalidTxnVersion
}

// NextVersionToExecute get the next transaction index to execute,
// returns invalid version if no task is available
//
// Invariant `num_active_tasks`: increased if a valid task is returned.
func (s *Scheduler) NextVersionToExecute() TxnVersion {
	if s.execution_idx.Load() >= uint64(s.block_size) {
		s.CheckDone()
		return InvalidTxnVersion
	}
	IncrAtomic(&s.num_active_tasks)
	idx_to_execute := s.execution_idx.Add(1) - 1
	return s.TryIncarnate(TxnIndex(idx_to_execute))
}

// NextVersionToValidate get the next transaction index to validate,
// returns invalid version if no task is available.
//
// Invariant `num_active_tasks`: increased if a valid task is returned.
func (s *Scheduler) NextVersionToValidate() TxnVersion {
	if s.validation_idx.Load() >= uint64(s.block_size) {
		s.CheckDone()
		return InvalidTxnVersion
	}
	IncrAtomic(&s.num_active_tasks)
	idx_to_validate := FetchIncr(&s.validation_idx)
	if idx_to_validate < uint64(s.block_size) {
		if ok, incarnation := s.txn_status[idx_to_validate].IsExecuted(); ok {
			return TxnVersion{TxnIndex(idx_to_validate), incarnation}
		}
	}

	DecrAtomic(&s.num_active_tasks)
	return InvalidTxnVersion
}

// NextTask returns the transaction index and task kind for the next task to execute or validate,
// returns invalid version if no task is available.
//
// Invariant `num_active_tasks`: increased if a valid task is returned.
func (s *Scheduler) NextTask() (TxnVersion, TaskKind) {
	validation_idx := s.validation_idx.Load()
	execution_idx := s.execution_idx.Load()
	if validation_idx < execution_idx {
		return s.NextVersionToValidate(), TaskKindValidation
	} else {
		return s.NextVersionToExecute(), TaskKindExecution
	}
}

func (s *Scheduler) WaitForDependency(txn TxnIndex, blocking_txn TxnIndex) *Condvar {
	cond := NewCondvar()
	entry := &s.txn_dependency[blocking_txn]
	entry.Lock()

	// thread holds 2 locks
	if ok, _ := s.txn_status[blocking_txn].IsExecuted(); ok {
		// dependency resolved before locking in Line 148
		entry.Unlock()
		return nil
	}

	s.txn_status[txn].Suspend(cond)
	entry.dependents = append(entry.dependents, txn)
	entry.Unlock()

	return cond
}

func (s *Scheduler) ResumeDependencies(txns []TxnIndex) {
	for _, txn := range txns {
		s.txn_status[txn].Resume()
	}
}

// Invariant `num_active_tasks`: decreased if an invalid task is returned.
func (s *Scheduler) FinishExecution(version TxnVersion, wroteNewPath bool) (TxnVersion, TaskKind) {
	s.txn_status[version.Index].SetExecuted()

	deps := s.txn_dependency[version.Index].Swap(nil)
	s.ResumeDependencies(deps)
	if s.validation_idx.Load() > uint64(version.Index) { // otherwise index already small enough
		if !wroteNewPath {
			// schedule validation for current tx only, don't decrease num_active_tasks
			return version, TaskKindValidation
		}
		// schedule validation for txn_idx and higher txns
		s.DecreaseValidationIdx(version.Index)
	}
	DecrAtomic(&s.num_active_tasks)
	return InvalidTxnVersion, 0
}

func (s *Scheduler) TryValidationAbort(version TxnVersion) bool {
	return s.txn_status[version.Index].TryValidationAbort(version.Incarnation)
}

// Invariant `num_active_tasks`: decreased if an invalid task is returned.
func (s *Scheduler) FinishValidation(txn TxnIndex, aborted bool) (TxnVersion, TaskKind) {
	if aborted {
		s.txn_status[txn].SetReadyStatus()
		s.DecreaseValidationIdx(txn + 1)
		if s.execution_idx.Load() > uint64(txn) {
			return s.TryIncarnate(txn), TaskKindExecution
		}
	}

	DecrAtomic(&s.num_active_tasks)
	return InvalidTxnVersion, 0
}

func (s *Scheduler) Stats() string {
	return fmt.Sprintf("executed: %d, validated: %d",
		s.executedTxns.Load(), s.validatedTxns.Load())
}
