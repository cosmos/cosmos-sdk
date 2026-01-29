package blockstm

import (
	"fmt"
	"sync"
)

type Status uint

const (
	StatusReadyToExecute Status = iota
	StatusExecuting
	StatusExecuted
	StatusAborting
	StatusSuspended
	StatusCommitted
)

// ExecutionStatus is a state machine for the status of a transaction, all the transitions are atomic protected by a mutex.
//
// ```mermaid
// stateDiagram-v2
//
//	[*] --> ReadyToExecute
//	ReadyToExecute --> Executing: TrySetExecuting()
//	Executing --> Executed: SetExecuted()
//	Executing --> Suspended: Suspend(cond)\nset cond
//	Executed --> Aborting: TryValidationAbort(incarnation)
//	Aborting --> ReadyToExecute: SetReadyStatus()\nincarnation++
//	Suspended --> Executing: Resume()
//
// ```
type ExecutionStatus struct {
	sync.Mutex

	incarnation Incarnation
	status      Status

	cond *Condvar
}

// If includeCommitted is true (which is when calling from WaitForDependency),
// then committed transaction is also considered executed (for dependency resolution
// purposes). If includeCommitted is false (which is when calling from
// NextVersionToValidate), then we are checking if a transaction may be validated,
// and a committed (in between) txn does not need to be scheduled for validation -
// so can return None.
func (s *ExecutionStatus) IsExecuted(includeCommitted bool) (incarnation Incarnation, ok bool) {
	s.Lock()

	if s.status == StatusExecuted || (includeCommitted && s.status == StatusCommitted) {
		ok = true
		incarnation = s.incarnation
	}

	s.Unlock()
	return incarnation, ok
}

func (s *ExecutionStatus) TryIsExecuted() (incarnation Incarnation, ok bool) {
	if !s.TryLock() {
		return 0, false
	}

	if s.status == StatusExecuted {
		ok = true
		incarnation = s.incarnation
	}

	s.Unlock()
	return incarnation, ok
}

// ExecutedOnce returns true iff the transaction has executed at least once,
// default to false if the lock cannot be acquired.
func (s *ExecutionStatus) ExecutedOnce() bool {
	if !s.TryLock() {
		return false
	}

	ok := s.incarnation > 0 ||
		s.status == StatusExecuted ||
		s.status == StatusAborting ||
		s.status == StatusCommitted

	s.Unlock()
	return ok
}

func (s *ExecutionStatus) TrySetExecuting() (incarnation Incarnation, ok bool) {
	s.Lock()

	if s.status == StatusReadyToExecute {
		s.status = StatusExecuting
		incarnation = s.incarnation
		ok = true
	}

	s.Unlock()
	return incarnation, ok
}

// setStatus sets the status to the given status if the current status is preStatus.
// preStatus invariant must be held by the caller.
func (s *ExecutionStatus) setStatus(status, preStatus Status) {
	s.Lock()

	if s.status != preStatus {
		s.Unlock()
		panic(fmt.Sprintf("invalid status transition: %v -> %v, current: %v", preStatus, status, s.status))
	}

	s.status = status
	s.Unlock()
}

func (s *ExecutionStatus) Resume() {
	s.Lock()

	// status must be SUSPENDED and cond != nil
	if s.status != StatusSuspended || s.cond == nil {
		s.Unlock()
		panic(fmt.Sprintf("invalid resume: status=%v", s.status))
	}

	s.status = StatusExecuting
	s.cond.Notify()
	s.cond = nil

	s.Unlock()
}

func (s *ExecutionStatus) SetExecuted() {
	// status must have been EXECUTING
	s.setStatus(StatusExecuted, StatusExecuting)
}

func (s *ExecutionStatus) TryValidationAbort(incarnation Incarnation) (ok bool) {
	s.Lock()

	if s.incarnation == incarnation && s.status == StatusExecuted {
		s.status = StatusAborting
		ok = true
	}

	s.Unlock()
	return ok
}

func (s *ExecutionStatus) SetReadyStatus() {
	s.Lock()

	// status must be ABORTING
	if s.status != StatusAborting {
		s.Unlock()
		panic(fmt.Sprintf("invalid status transition: %v -> %v, current: %v", StatusAborting, StatusReadyToExecute, s.status))
	}

	s.incarnation++
	s.status = StatusReadyToExecute

	s.Unlock()
}

func (s *ExecutionStatus) Suspend(cond *Condvar) {
	s.Lock()

	if s.status != StatusExecuting {
		s.Unlock()
		panic(fmt.Sprintf("invalid suspend: status=%v", s.status))
	}

	s.cond = cond
	s.status = StatusSuspended

	s.Unlock()
}

func (s *ExecutionStatus) SetCommitted() {
	s.setStatus(StatusCommitted, StatusExecuted)
}
