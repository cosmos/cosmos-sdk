package blockstm

import (
	"fmt"
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

// StatusEntry contains both execution and validation status for a transaction
type StatusEntry struct {
	execution  ExecutionStatus
	validation ValidationStatus
}

type CommitState struct {
	// next transaction to commit
	Index TxnIndex
	// sweeping lower bound on the wave of a validation that must
	// be successful in order to commit the next transaction.
	// it's the maximum triggered wave in transactions [0, Index].
	Wave Wave
}

// Scheduler implements the scheduler for the block-stm
// ref: `Algorithm 4 The Scheduler module, variables, utility APIs and next task logic`
type Scheduler struct {
	blockSize int

	// An index that tracks the next transaction to try and execute.
	executionIdx atomic.Uint32
	// A similar index for tracking validation.
	validationIdx atomic.Uint64
	// Marker for completion
	doneMarker atomic.Bool

	commitLock  ArmedLock
	commitState CommitState // commitState is only accessed when acquired commitLock

	// txnIdx to a mutex-protected set of dependent transaction indices
	txnDependency []TxDependency
	// txnIdx to a mutex-protected pair (incarnationNumber, status), where status ∈ {READY_TO_EXECUTE, EXECUTING, EXECUTED, ABORTING, SUSPENDED}.
	txnStatus []StatusEntry

	// metrics
	executedTxns  atomic.Int64
	validatedTxns atomic.Int64
}

func NewScheduler(blockSize int) *Scheduler {
	s := &Scheduler{
		blockSize:     blockSize,
		txnDependency: make([]TxDependency, blockSize),
		txnStatus:     make([]StatusEntry, blockSize),
	}
	s.commitLock.Init()
	for i := 0; i < blockSize; i++ {
		s.txnStatus[i].validation.Init()
	}
	return s
}

func (s *Scheduler) Done() bool {
	return s.doneMarker.Load()
}

func (s *Scheduler) DecreaseValidationIdx(target TxnIndex) (Wave, bool) {
	if int(target) == s.blockSize {
		// DecreaseValidationIdx is called with `txn + 1`, so it can equal blockSize.
		return 0, false
	}

	prev, ok := FetchUpdate(&s.validationIdx, func(current uint64) (uint64, bool) {
		txnIdx, wave := UnpackValidationIdx(current)
		if txnIdx > target {
			s.txnStatus[target].validation.SetTriggeredWave(wave + 1)
			return PackValidationIdx(target, wave+1), true
		}
		return current, false
	})
	if !ok {
		return 0, false
	}
	_, wave := UnpackValidationIdx(prev)
	return wave + 1, true
}

// TryIncarnate tries to incarnate a transaction index to execute.
// Returns the transaction version if successful, otherwise returns invalid version.
func (s *Scheduler) TryIncarnate(idx TxnIndex) TxnVersion {
	if idx < TxnIndex(s.blockSize) {
		if incarnation, ok := s.txnStatus[idx].execution.TrySetExecuting(); ok {
			return TxnVersion{idx, incarnation}
		}
	}
	return InvalidTxnVersion
}

// NextVersionToExecute get the next transaction index to execute,
// returns invalid version if no task is available
func (s *Scheduler) NextVersionToExecute() TxnVersion {
	if int(s.executionIdx.Load()) >= s.blockSize {
		return InvalidTxnVersion
	}
	idxToExecute := s.executionIdx.Add(1) - 1
	return s.TryIncarnate(TxnIndex(idxToExecute))
}

// TryValidateNextVersion get the next transaction index to validate,
// returns invalid version if no task is available.
func (s *Scheduler) TryValidateNextVersion(idxToValidate TxnIndex, wave Wave) (TxnVersion, bool) {
	old := PackValidationIdx(idxToValidate, wave)
	new := PackValidationIdx(idxToValidate+1, wave)
	if !s.validationIdx.CompareAndSwap(old, new) {
		return InvalidTxnVersion, false
	}

	incarnation, ok := s.txnStatus[idxToValidate].execution.IsExecuted(false)
	if !ok {
		return InvalidTxnVersion, false
	}

	return TxnVersion{TxnIndex(idxToValidate), incarnation}, true
}

// NextTask returns the transaction index and task kind for the next task to execute or validate,
// returns invalid version if no task is available.
func (s *Scheduler) NextTask() (TxnVersion, Wave, TaskKind) {
	validationIdx, wave := UnpackValidationIdx(s.validationIdx.Load())
	executionIdx := TxnIndex(s.executionIdx.Load())

	preferValidate := validationIdx < min(executionIdx, TxnIndex(s.blockSize)) &&
		s.txnStatus[validationIdx].execution.ExecutedOnce()

	if preferValidate {
		if version, ok := s.TryValidateNextVersion(validationIdx, wave); ok {
			return version, wave, TaskKindValidation
		}
	}

	return s.NextVersionToExecute(), 0, TaskKindExecution
}

func (s *Scheduler) WaitForDependency(txn, blockingTxn TxnIndex) *Condvar {
	cond := NewCondvar()
	entry := &s.txnDependency[blockingTxn]
	entry.Lock()

	// thread holds 2 locks
	if _, ok := s.txnStatus[blockingTxn].execution.IsExecuted(true); ok {
		// dependency resolved before entry.Lock() (https://github.com/cosmos/cosmos-sdk/blob/825fd620889acac4d0fd1bf0f9370651d2ee6610/blockstm/scheduler.go#L152) was acquired
		entry.Unlock()
		return nil
	}

	s.txnStatus[txn].execution.Suspend(cond)
	entry.dependents = append(entry.dependents, txn)
	entry.Unlock()

	return cond
}

func (s *Scheduler) ResumeDependencies(txns []TxnIndex) {
	for _, txn := range txns {
		s.txnStatus[txn].execution.Resume()
	}
}

// FinishExecution marks an execution task as complete.
func (s *Scheduler) FinishExecution(version TxnVersion, wroteNewPath bool) (TxnVersion, Wave, TaskKind) {
	// grab validation status lock before execution status is changed
	validationStatus := &s.txnStatus[version.Index].validation
	validationStatus.Lock()
	defer validationStatus.Unlock()

	s.txnStatus[version.Index].execution.SetExecuted()

	deps := s.txnDependency[version.Index].Swap(nil)
	s.ResumeDependencies(deps)
	validationIdx, wave := UnpackValidationIdx(s.validationIdx.Load())
	if validationIdx > TxnIndex(version.Index) { // otherwise index already small enough
		if wroteNewPath {
			// schedule validation wave for higher txns
			curWave, ok := s.DecreaseValidationIdx(version.Index + 1)
			if ok {
				wave = curWave
			}
		}

		// schedule specific validation for current tx
		validationStatus.RequiredWave = wave
		return version, wave, TaskKindValidation
	}
	return InvalidTxnVersion, 0, 0
}

func (s *Scheduler) TryValidationAbort(version TxnVersion) bool {
	return s.txnStatus[version.Index].execution.TryValidationAbort(version.Incarnation)
}

// FinishValidation marks a validation task as complete.
func (s *Scheduler) FinishValidation(txn TxnIndex, wave Wave, aborted bool, valid bool) (TxnVersion, TaskKind) {
	if aborted {
		s.txnStatus[txn].execution.SetReadyStatus()
		s.DecreaseValidationIdx(txn + 1)
		if TxnIndex(s.executionIdx.Load()) > txn {
			return s.TryIncarnate(txn), TaskKindExecution
		}
	}

	if valid {
		// process validation wave
		s.txnStatus[txn].validation.SetValidatedWave(wave)
		// mark as commitable
		s.commitLock.Arm()
	}

	return InvalidTxnVersion, 0
}

// TryCommit is called with commitLock held, no concurrency.
func (s *Scheduler) TryCommit() (TxnIndex, Incarnation, bool) {
	commitIdx := s.commitState.Index
	if commitIdx == TxnIndex(s.blockSize) {
		return 0, 0, false
	}

	validationStatus := &s.txnStatus[commitIdx].validation
	if !validationStatus.TryLock() {
		return 0, 0, false
	}
	defer validationStatus.Unlock()

	execStatus := &s.txnStatus[commitIdx].execution
	if !execStatus.TryLock() {
		return 0, 0, false
	}
	defer execStatus.Unlock()

	if execStatus.status != StatusExecuted {
		// not yet executed
		return 0, 0, false
	}

	// update commit wave with triggered wave of this txn
	s.commitState.Wave = max(s.commitState.Wave, validationStatus.TriggeredWave)

	// check if validated wave is late enough
	if validationStatus.ValidatedWave < 0 {
		// not yet validated
		return 0, 0, false
	}
	if Wave(validationStatus.ValidatedWave) < max(s.commitState.Wave, validationStatus.RequiredWave) {
		// not late enough
		return 0, 0, false
	}

	execStatus.status = StatusCommitted

	s.commitState.Index++
	if s.commitState.Index == TxnIndex(s.blockSize) {
		s.doneMarker.Store(true)
	}

	return commitIdx, execStatus.incarnation, true
}

func (s *Scheduler) ProcessCommits() {
	// keep processing if there's work to do and we can acquire the lock
	for s.commitLock.TryLock() {
		for {
			_, _, ok := s.TryCommit()
			if !ok {
				break
			}

			// TODO commit handling
		}

		s.commitLock.Unlock()
	}
}

func (s *Scheduler) Stats() string {
	return fmt.Sprintf("executed: %d, validated: %d",
		s.executedTxns.Load(), s.validatedTxns.Load())
}
