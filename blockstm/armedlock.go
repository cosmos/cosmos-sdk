package blockstm

import "sync/atomic"

type ArmedLock struct {
	// Last bit:   1 -> unlocked; 0 -> locked
	// Second bit: 1 -> there's work; 0 -> no work
	lock atomic.Uint64
}

// Init set the lock to unlocked and armed (with work to do).
func (l *ArmedLock) Init() {
	l.lock.Store(0x11)
}

// TryLock returns true if there's work to do and the lock is acquired.
func (l *ArmedLock) TryLock() bool {
	return l.lock.CompareAndSwap(0x11, 0x00)
}

func (l *ArmedLock) Unlock() {
	l.lock.Or(0x01)
}

func (l *ArmedLock) Arm() {
	l.lock.Or(0x10)
}
