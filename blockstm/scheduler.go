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
	blockSize int

	// An index that tracks the next transaction to try and execute.
	executionIdx atomic.Uint64
	// A similar index for tracking validation.
	validationIdx atomic.Uint64
	// Number of times validationIdx or executionIdx was decreased
	decreaseCnt atomic.Uint64
	// Number of ongoing validation and execution tasks
	numActiveTasks atomic.Uint64
	// Marker for completion
	doneMarker atomic.Bool

	// txnIdx to a mutex-protected set of dependent transaction indices
	txnDependency []TxDependency
	// txnIdx to a mutex-protected pair (incarnationNumber, status), where status ∈ {READY_TO_EXECUTE, EXECUTING, EXECUTED, ABORTING}.
	txnStatus []StatusEntry

	// metrics
	executedTxns  atomic.Int64
	validatedTxns atomic.Int64
}

func NewScheduler(blockSize int) *Scheduler {
	return &Scheduler{
		blockSize:     blockSize,
		txnDependency: make([]TxDependency, blockSize),
		txnStatus:     make([]StatusEntry, blockSize),
	}
}

func (s *Scheduler) Done() bool {
	return s.doneMarker.Load()
}

func (s *Scheduler) DecreaseValidationIdx(target TxnIndex) {
	StoreMin(&s.validationIdx, uint64(target))
	s.decreaseCnt.Add(1)
}

func (s *Scheduler) CheckDone() {
	observedCnt := s.decreaseCnt.Load()
	if s.executionIdx.Load() >= uint64(s.blockSize) &&
		s.validationIdx.Load() >= uint64(s.blockSize) &&
		s.numActiveTasks.Load() == 0 {
		if observedCnt == s.decreaseCnt.Load() {
			s.doneMarker.Store(true)
		}
	}
	// avoid busy waiting
	runtime.Gosched()
}

// TryIncarnate tries to incarnate a transaction index to execute.
// Returns the transaction version if successful, otherwise returns invalid version.
//
// Invariant `numActiveTasks`: decreased if an invalid task is returned.
func (s *Scheduler) TryIncarnate(idx TxnIndex) TxnVersion {
	if int(idx) < s.blockSize {
		if incarnation, ok := s.txnStatus[idx].TrySetExecuting(); ok {
			return TxnVersion{idx, incarnation}
		}
	}
	DecrAtomic(&s.numActiveTasks)
	return InvalidTxnVersion
}

// NextVersionToExecute get the next transaction index to execute,
// returns invalid version if no task is available
//
// Invariant `numActiveTasks`: increased if a valid task is returned.
func (s *Scheduler) NextVersionToExecute() TxnVersion {
	if s.executionIdx.Load() >= uint64(s.blockSize) {
		s.CheckDone()
		return InvalidTxnVersion
	}
	IncrAtomic(&s.numActiveTasks)
	idxToExecute := s.executionIdx.Add(1) - 1
	return s.TryIncarnate(TxnIndex(idxToExecute))
}

// NextVersionToValidate get the next transaction index to validate,
// returns invalid version if no task is available.
//
// Invariant `numActiveTasks`: increased if a valid task is returned.
func (s *Scheduler) NextVersionToValidate() TxnVersion {
	if s.validationIdx.Load() >= uint64(s.blockSize) {
		s.CheckDone()
		return InvalidTxnVersion
	}
	IncrAtomic(&s.numActiveTasks)
	idxToValidate := FetchIncr(&s.validationIdx)
	if idxToValidate < uint64(s.blockSize) {
		if ok, incarnation := s.txnStatus[idxToValidate].IsExecuted(); ok {
			return TxnVersion{TxnIndex(idxToValidate), incarnation}
		}
	}

	DecrAtomic(&s.numActiveTasks)
	return InvalidTxnVersion
}

// NextTask returns the transaction index and task kind for the next task to execute or validate,
// returns invalid version if no task is available.
//
// Invariant `numActiveTasks`: increased if a valid task is returned.
func (s *Scheduler) NextTask() (TxnVersion, TaskKind) {
	validationIdx := s.validationIdx.Load()
	executionIdx := s.executionIdx.Load()
	if validationIdx < executionIdx {
		return s.NextVersionToValidate(), TaskKindValidation
	} else {
		return s.NextVersionToExecute(), TaskKindExecution
	}
}

func (s *Scheduler) WaitForDependency(txn, blockingTxn TxnIndex) *Condvar {
	cond := NewCondvar()
	entry := &s.txnDependency[blockingTxn]
	entry.Lock()

	// thread holds 2 locks
	if ok, _ := s.txnStatus[blockingTxn].IsExecuted(); ok {
		// dependency resolved before locking in Line 148
		entry.Unlock()
		return nil
	}

	s.txnStatus[txn].Suspend(cond)
	entry.dependents = append(entry.dependents, txn)
	entry.Unlock()

	return cond
}

func (s *Scheduler) ResumeDependencies(txns []TxnIndex) {
	for _, txn := range txns {
		s.txnStatus[txn].Resume()
	}
}

// FinishExecution marks an execution task as complete.
// Invariant `numActiveTasks`: decreased if an invalid task is returned.
func (s *Scheduler) FinishExecution(version TxnVersion, wroteNewPath bool) (TxnVersion, TaskKind) {
	s.txnStatus[version.Index].SetExecuted()

	deps := s.txnDependency[version.Index].Swap(nil)
	s.ResumeDependencies(deps)
	if s.validationIdx.Load() > uint64(version.Index) { // otherwise index already small enough
		if !wroteNewPath {
			// schedule validation for current tx only, don't decrease numActiveTasks
			return version, TaskKindValidation
		}
		// schedule validation for txnIdx and higher txns
		s.DecreaseValidationIdx(version.Index)
	}
	DecrAtomic(&s.numActiveTasks)
	return InvalidTxnVersion, 0
}

func (s *Scheduler) TryValidationAbort(version TxnVersion) bool {
	return s.txnStatus[version.Index].TryValidationAbort(version.Incarnation)
}

// FinishValidation marks a validation task as complete.
// Invariant `numActiveTasks`: decreased if an invalid task is returned.
func (s *Scheduler) FinishValidation(txn TxnIndex, aborted bool) (TxnVersion, TaskKind) {
	if aborted {
		s.txnStatus[txn].SetReadyStatus()
		s.DecreaseValidationIdx(txn + 1)
		if s.executionIdx.Load() > uint64(txn) {
			return s.TryIncarnate(txn), TaskKindExecution
		}
	}

	DecrAtomic(&s.numActiveTasks)
	return InvalidTxnVersion, 0
}

func (s *Scheduler) Stats() string {
	return fmt.Sprintf("executed: %d, validated: %d",
		s.executedTxns.Load(), s.validatedTxns.Load())
}
