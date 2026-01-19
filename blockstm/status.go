package blockstm

import "sync"

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

func (s *StatusEntry) setStatus(status Status) {
	s.Lock()
	s.status = status
	s.Unlock()
}

func (s *StatusEntry) Resume() {
	// status must be SUSPENDED and cond != nil
	s.Lock()

	s.status = StatusExecuting
	s.cond.Notify()
	s.cond = nil

	s.Unlock()
}

func (s *StatusEntry) SetExecuted() {
	// status must have been EXECUTING
	s.setStatus(StatusExecuted)
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

	s.incarnation++
	// status must be ABORTING
	s.status = StatusReadyToExecute

	s.Unlock()
}

func (s *StatusEntry) Suspend(cond *Condvar) {
	s.Lock()

	s.cond = cond
	s.status = StatusSuspended

	s.Unlock()
}

// TryCancel wakes up a suspended executor if it's suspended.
// Called during context cancellation to prevent hanging.
// Returns true if the transaction was suspended and is now canceled.
func (s *StatusEntry) TryCancel() bool {
	s.Lock()
	defer s.Unlock()

	if s.status == StatusSuspended && s.cond != nil {
		s.status = StatusExecuting
		s.cond.Notify()
		s.cond = nil
		return true
	}
	return false
}
