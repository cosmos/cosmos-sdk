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
)

// StatusEntry is a state machine for the status of a transaction, all the transitions are atomic protected by a mutex.
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
type StatusEntry struct {
	sync.Mutex

	incarnation Incarnation
	status      Status

	cond *Condvar
}

func (s *StatusEntry) IsExecuted() (incarnation Incarnation, ok bool) {
	s.Lock()

	if s.status == StatusExecuted {
		ok = true
		incarnation = s.incarnation
	}

	s.Unlock()
	return incarnation, ok
}

// ExecutedOnce returns true iff the transaction has executed at least once,
// default to false if the lock cannot be acquired.
func (s *StatusEntry) ExecutedOnce() bool {
	if !s.TryLock() {
		return false
	}

	ok := s.incarnation > 0 ||
		s.status == StatusExecuted ||
		s.status == StatusAborting

	s.Unlock()
	return ok
}

func (s *StatusEntry) TrySetExecuting() (incarnation Incarnation, ok bool) {
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
func (s *StatusEntry) setStatus(status, preStatus Status) {
	s.Lock()

	if s.status != preStatus {
		s.Unlock()
		panic(fmt.Sprintf("invalid status transition: %v -> %v, current: %v", preStatus, status, s.status))
	}

	s.status = status
	s.Unlock()
}

func (s *StatusEntry) Resume() {
	// Resume is normally called for a txn that is currently suspended.
	// With cancellation, a suspended txn may already have been woken and had its
	// condition cleared; in that case this becomes a no-op.
	s.Lock()
	defer s.Unlock()

	if s.status != StatusSuspended || s.cond == nil {
		return
	}

	// status must be SUSPENDED and cond != nil
	if s.status != StatusSuspended || s.cond == nil {
		s.Unlock()
		panic(fmt.Sprintf("invalid resume: status=%v", s.status))
	}

	// status must be SUSPENDED and cond != nil
	if s.status != StatusSuspended || s.cond == nil {
		s.Unlock()
		panic(fmt.Sprintf("invalid resume: status=%v", s.status))
	}

	s.status = StatusExecuting
	s.cond.Notify()
	s.cond = nil
}

func (s *StatusEntry) SetExecuted() {
	// status must have been EXECUTING
	s.setStatus(StatusExecuted, StatusExecuting)
}

func (s *StatusEntry) TryValidationAbort(incarnation Incarnation) (ok bool) {
	s.Lock()

	if s.incarnation == incarnation && s.status == StatusExecuted {
		s.status = StatusAborting
		ok = true
	}

	s.Unlock()
	return ok
}

func (s *StatusEntry) SetReadyStatus() {
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

func (s *StatusEntry) Suspend(cond *Condvar) {
	s.Lock()

	if s.status != StatusExecuting {
		s.Unlock()
		panic(fmt.Sprintf("invalid suspend: status=%v", s.status))
	}

	s.cond = cond
	s.status = StatusSuspended

	s.Unlock()
}

// TryCancel wakes up a suspended executor if it's suspended.
// Called during context cancellation to prevent hanging.
func (s *StatusEntry) TryCancel(preCancel func()) {
	s.Lock()
	defer s.Unlock()

	if s.status == StatusSuspended {
		if preCancel != nil {
			preCancel()
		}

		if s.cond != nil {
			s.status = StatusExecuting
			s.cond.Notify()
			s.cond = nil
		}
	}
}
